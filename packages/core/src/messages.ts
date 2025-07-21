import { z } from "zod";

export const MessageRole = z.enum(["system", "user", "assistant", "tool"]);
export type MessageRole = z.infer<typeof MessageRole>;

export const ToolCall = z.object({
  id: z.string(),
  name: z.string(),
  parameters: z.record(z.any()),
});
export type ToolCall = z.infer<typeof ToolCall>;

export const ToolResult = z.object({
  id: z.string(),
  result: z.any(),
  error: z.string().optional(),
});
export type ToolResult = z.infer<typeof ToolResult>;
export const Message = z.object({
  id: z.string(),
  role: MessageRole,
  content: z.string(),
  toolCalls: z.array(ToolCall).optional(),
  toolResults: z.array(ToolResult).optional(),
  timestamp: z.number(),
  tokens: z
    .object({
      input: z.number().optional(),
      output: z.number().optional(),
    })
    .optional(),
});
export type Message = z.infer<typeof Message>;

export const Conversation = z.object({
  id: z.string(),
  messages: z.array(Message),
  totalTokens: z.object({
    input: z.number(),
    output: z.number(),
  }),
  createdAt: z.number(),
  updatedAt: z.number(),
});
export type Conversation = z.infer<typeof Conversation>;

export function createMessage(
  role: MessageRole,
  content: string,
  options: Partial<Pick<Message, "toolCalls" | "toolResults" | "tokens">> = {},
): Message {
  return {
    id: crypto.randomUUID(),
    role,
    content,
    timestamp: Date.now(),
    ...options,
  };
}

export function createConversation(): Conversation {
  return {
    id: crypto.randomUUID(),
    messages: [],
    totalTokens: { input: 0, output: 0 },
    createdAt: Date.now(),
    updatedAt: Date.now(),
  };
}

if (import.meta.main) {
  const testConversation = createConversation();
  const message = createMessage("user", "Hello!");

  console.log("Test conversation:", testConversation);
  console.log("Test message:", message);
}
