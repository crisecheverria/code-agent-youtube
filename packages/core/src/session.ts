import { z } from "zod";

export const SessionConfig = z.object({
  model: z.string().default("llama-3.3-70b-versatile"),
});

export type SessionConfig = z.infer<typeof SessionConfig>;

// Simple conversation interface for now
export interface Conversation {
  id: string;
  createdAt: number;
  updatedAt: number;
  messages: any[]; // We'll properly type this in episode 2
}

// Basic Session class for step 1
export class Session {
  private conversation: Conversation;
  private config: SessionConfig;

  constructor(config: SessionConfig) {
    this.config = SessionConfig.parse(config);
    this.conversation = {
      id: crypto.randomUUID(),
      createdAt: Date.now(),
      updatedAt: Date.now(),
      messages: [],
    };
  }

  getConversation(): Conversation {
    return { ...this.conversation };
  }

  getConfig(): SessionConfig {
    return this.config;
  }
}
