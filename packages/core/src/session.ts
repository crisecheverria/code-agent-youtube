import { z } from "zod";
import {
  Conversation,
  createConversation,
  createMessage,
  Message,
} from "./messages";
import {
  bashTool,
  listFilesTool,
  makeDirTool,
  readFileTool,
  ToolExecutor,
  writeFileTool,
} from "./tools";
import { GroqClient } from "./groq";

export const SessionConfig = z.object({
  groq: z.object({
    token: z.string(),
    model: z.string().default("llama-3.3-70b-versatile"),
    baseURL: z.string().default("https://api.groq.com/openai"),
  }),
});

export type SessionConfig = z.infer<typeof SessionConfig>;

export class Session {
  private conversation: Conversation;
  private groq: GroqClient;
  private toolExecutor: ToolExecutor;

  constructor(config: SessionConfig) {
    const validatedConfig = SessionConfig.parse(config);

    this.conversation = createConversation();
    this.groq = new GroqClient(validatedConfig.groq);
    this.toolExecutor = new ToolExecutor();

    // Register built-in tools
    this.toolExecutor.registerTool(bashTool);
    this.toolExecutor.registerTool(readFileTool);
    this.toolExecutor.registerTool(writeFileTool);
    this.toolExecutor.registerTool(listFilesTool);
    this.toolExecutor.registerTool(makeDirTool);

    // Add system prompt
    const systemMessage = createMessage(
      "system",
      "You are a helpful AI coding assistant. You can answer questions directly and use tools when needed. For simple questions like math problems or general knowledge, answer directly without using tools. Use tools only when you need to interact with the file system, execute code, or perform system operations.",
    );
    this.conversation.messages.push(systemMessage);
  }

  async sendMessage(content: string): Promise<Message> {
    // Add user message to conversation
    const userMessage = createMessage("user", content);
    this.conversation.messages.push(userMessage);

    // Get available tools
    const tools = this.toolExecutor.getGroqAITools();

    // Get response from Groq
    const response = await this.groq.complete(
      this.conversation.messages,
      tools,
    );

    // Handle tool calls
    if (response.toolCalls && response.toolCalls.length > 0) {
      // Create assistant message with tool calls
      const assistantMessage = createMessage(
        "assistant",
        response.content || "",
        {
          tokens: response.tokens,
          toolCalls: response.toolCalls.map((call) => ({
            id: call.id,
            name: call.function.name,
            parameters: JSON.parse(call.function.arguments),
          })),
        },
      );

      this.conversation.messages.push(assistantMessage);

      // Execute tool calls and add results
      for (const toolCall of response.toolCalls) {
        try {
          const params = JSON.parse(toolCall.function.arguments);
          const execution = await this.toolExecutor.execute(
            toolCall.function.name,
            params,
          );

          // Add tool result message
          const toolMessage = createMessage(
            "tool",
            JSON.stringify(execution.output),
            {
              toolResults: [
                {
                  id: toolCall.id,
                  result: execution.output,
                  error: execution.error,
                },
              ],
            },
          );

          this.conversation.messages.push(toolMessage);
        } catch (error) {
          // Add error message if tool execution fails
          const errorMessage = createMessage(
            "tool",
            JSON.stringify({
              error: error instanceof Error ? error.message : String(error),
            }),
            {
              toolResults: [
                {
                  id: toolCall.id,
                  result: null,
                  error: error instanceof Error ? error.message : String(error),
                },
              ],
            },
          );

          this.conversation.messages.push(errorMessage);
        }
      }
      // Get final response from Groq
      const finalResponse = await this.groq.complete(
        this.conversation.messages,
        tools,
      );
      const finalMessage = createMessage(
        "assistant",
        finalResponse.content || "",
        {
          tokens: finalResponse.tokens,
        },
      );

      this.conversation.messages.push(finalMessage);

      // Update conversation total tokens
      this.conversation.totalTokens.input +=
        response.tokens.input + finalResponse.tokens.input;
      this.conversation.totalTokens.output +=
        response.tokens.output + finalResponse.tokens.output;
      this.conversation.updatedAt = Date.now();

      return finalMessage;
    } else {
      // No tool calls, just regular response
      const assistantMessage = createMessage("assistant", response.content, {
        tokens: response.tokens,
      });

      this.conversation.messages.push(assistantMessage);

      // Update conversation token counts
      this.conversation.totalTokens.input += response.tokens.input;
      this.conversation.totalTokens.output += response.tokens.output;
      this.conversation.updatedAt = Date.now();

      return assistantMessage;
    }
  }

  async *streamMessage(
    content: string,
  ): AsyncGenerator<string, Message, unknown> {
    // Add user message to conversation
    const userMessage = createMessage("user", content);
    this.conversation.messages.push(userMessage);

    let assistantContent = "";

    // Stream response from Groq
    const stream = await this.groq.stream(this.conversation.messages);

    for await (const chunk of stream) {
      assistantContent += chunk;
      yield chunk;
    }

    // Create assistant message
    const assistantMessage = createMessage("assistant", assistantContent);
    this.conversation.messages.push(assistantMessage);
    this.conversation.updatedAt = Date.now();

    return assistantMessage;
  }

  async executeTool(name: string, params: any): Promise<any> {
    const execution = await this.toolExecutor.execute(name, params);

    // Add tool result message to conversation
    const toolMessage = createMessage(
      "tool",
      JSON.stringify(execution.output),
      {
        toolResults: [
          {
            id: execution.id,
            result: execution.output,
            error: execution.error,
          },
        ],
      },
    );

    this.conversation.messages.push(toolMessage);
    this.conversation.updatedAt = Date.now();

    return execution;
  }

  getConversation(): Conversation {
    return { ...this.conversation };
  }

  getAvailableTools(): string[] {
    return this.toolExecutor.getTools();
  }

  getTokenUsage(): { input: number; output: number; total: number } {
    const { input, output } = this.conversation.totalTokens;
    return { input, output, total: input + output };
  }

  clear(): void {
    this.conversation = createConversation();
  }
}
