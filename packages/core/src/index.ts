import { Hono } from "hono";
import { Session, type SessionConfig } from "./session";
import { serve } from "bun";

const app = new Hono();

// Global session
let currentSession: Session | null = null;

// Health check endpoin
app.get("/health", (c) => {
  return c.json({
    status: "ok",
    timestamp: Date.now(),
    hasSession: !!currentSession,
  });
});

// Initialize session
app.post("/session", async (c) => {
  try {
    const config = (await c.req.json()) as SessionConfig;
    currentSession = new Session(config);
    return c.json({
      success: true,
      sessionId: currentSession.getConversation().id,
    });
  } catch (error) {
    return c.json(
      { success: false, error: "Failed to initialize session" },
      400,
    );
  }
});

const port = process.env.PORT ? parseInt(process.env.PORT) : 3000;

console.log(`🚀 Code Agent server starting on port ${port}`);

serve({
  fetch: app.fetch,
  port,
});

export { app };
