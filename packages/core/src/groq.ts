import { z } from "zod";
import type { Message } from "./messages.ts";

// Groq configuration
export const GroqConfig = z.object({
  token: z.string(),
  model: z.string().default("llama-3.3-70b-versatile"),
  baseURL: z.string().default("https://api.groq.com/openai"),
});
export type GroqConfig = z.infer<typeof GroqConfig>;

// Groq API response
export const GroqResponse = z.object({
  content: z.string(),
  tokens: z.object({
    input: z.number(),
    output: z.number(),
  }),
  toolCalls: z
    .array(
      z.object({
        id: z.string(),
        type: z.string(),
        function: z.object({
          name: z.string(),
          arguments: z.string(),
        }),
      }),
    )
    .optional(),
});
export type GroqResponse = z.infer<typeof GroqResponse>;

/**
 * GroqClient - Handles communication with Groq API
 */
export class GroqClient {
  private config: GroqConfig;

  constructor(config: GroqConfig) {
    if (!config) {
      throw new Error("GroqConfig is required");
    }
    if (!config.token) {
      throw new Error("Groq API token is required");
    }

    // Parse and validate config to apply default
    this.config = GroqConfig.parse(config);
  }

  async complete(messages: Message[], tools?: any[]): Promise<GroqResponse> {
    const payload: any = {
      model: this.config.model,
      messages: messages.map((msg) => {
        const groqMsg: any = {
          role: msg.role,
          content: msg.content,
        };

        // Handle tools calls in assistant messages
        if (msg.toolCalls && msg.toolCalls.length > 0) {
          groqMsg.tool_calls = msg.toolCalls.map((tool) => ({
            id: tool.id,
            type: "function",
            function: {
              name: tool.name,
              arguments: JSON.stringify(tool.parameters),
            },
          }));
        }

        // Handle tool results in user messages
        if (msg.toolResults && msg.toolResults.length > 0) {
          groqMsg.tool_call_id = msg.toolResults[0].id;
        }

        return groqMsg;
      }),
      stream: false,
      temperature: 0.7,
      max_tokens: 4096,
    };

    if (tools && tools.length > 0) {
      payload.tools = tools;
      payload.tool_choice = "auto";
    }

    //Retry logic for rate limits and temporary errors
    const maxRetries = 3;
    let lastError: Error | null = null;

    for (let attempt = 1; attempt < maxRetries; attempt++) {
      try {
        const response = await fetch(
          `${this.config.baseURL}/v1/chat/completions`,
          {
            method: "POST",
            headers: {
              Authorization: `Bearer ${this.config.token}`,
              "Content-Type": "application/json",
              "User-Agent": "code-agent/0.1.0",
            },
            body: JSON.stringify(payload),
          },
        );

        if (!response.ok) {
          const errorText = await response.text();
          const error = new Error(
            `Groq API error: ${response.status} ${response.statusText} - ${errorText}`,
          );

          // Check if it's a retryable error
          if (response.status === 429 || response.status >= 500) {
            lastError = error;
            if (attempt < maxRetries) {
              const delay = Math.pow(2, attempt) * 1000; // Exponential backoff
              console.warn(
                `Attempt ${attempt} failed, retrying in ${delay}ms...`,
              );
              await new Promise((resolve) => setTimeout(resolve, delay));
              continue;
            }
          }

          throw error;
        }

        const data = await response.json();

        const choice = data.choices[0];
        return {
          content: choice?.message?.content || "",
          tokens: {
            input: data.usage?.input_tokens || 0,
            output: data.usage?.output_tokens || 0,
          },
          toolCalls: choice?.message?.tool_calls || [],
        };
      } catch (error) {
        lastError = error instanceof Error ? error : new Error(String(error));
        if (attempt < maxRetries) {
          const delay = Math.pow(2, attempt) * 1000; // Exponential backoff
          console.warn(`Attempt ${attempt} failed, retrying in ${delay}ms...`);
          await new Promise((resolve) => setTimeout(resolve, delay));
          continue;
        }
        break;
      }
    }
    throw new Error(
      `Failed to complete with Groq after ${maxRetries} attempts: ${lastError?.message || "Unknown error"}`,
    );
  }

  async stream(
    messages: Message[],
  ): Promise<AsyncGenerator<string, void, unknown>> {
    const payload = {
      model: this.config.model,
      messages: messages.map((msg) => ({
        role: msg.role,
        content: msg.content,
      })),
      stream: true,
      temperature: 0.7,
      max_tokens: 4096,
    };

    const response = await fetch(`${this.config.baseURL}/v1/chat/completions`, {
      method: "POST",
      headers: {
        Authorization: `Bearer ${this.config.token}`,
        "Content-Type": "application/json",
        "User-Agent": "code-agent/0.1.0",
      },
      body: JSON.stringify(payload),
    });

    if (!response.ok) {
      const errorText = await response.text();
      throw new Error(
        `Groq API error: ${response.status} ${response.statusText} - ${errorText}`,
      );
    }

    return this.parseStream(response);
  }

  private async *parseStream(
    response: Response,
  ): AsyncGenerator<string, void, unknown> {
    const reader = response.body?.getReader();
    if (!reader) {
      throw new Error("Response body is not readable");
    }

    const decoder = new TextDecoder();
    let buffer = "";

    try {
      while (true) {
        const { done, value } = await reader.read();
        if (done) break;

        buffer += decoder.decode(value, { stream: true });
        const lines = buffer.split("\n");
        buffer = lines.pop() || ""; // Keep the last incomplete line in buffer

        for (const line of lines) {
          if (line.startsWith("data")) {
            const data = line.slice(6);
            if (data === "[DONE]") return;

            try {
              const parsed = JSON.parse(data);
              const content = parsed.choices[0]?.delta?.content || "";
              if (content) {
                yield content;
              }
            } catch (e) {
              // Skip invalid JSON lines
            }
          }
        }
      }
    } finally {
      reader.releaseLock();
    }
  }
}
