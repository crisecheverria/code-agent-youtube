import { serve } from "bun";
import { Hono } from "hono";
import { Session, type SessionConfig } from "./session";

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

// Send message
app.post("/message", async (c) => {
	if (!currentSession) {
		return c.json({ success: false, error: "No active session" }, 400);
	}

	try {
		const { content } = await c.req.json();
		const message = await currentSession.sendMessage(content);
		return c.json({ success: true, messages: [message] });
	} catch (error) {
		return c.json(
			{
				success: false,
				error: error instanceof Error ? error.message : "Unknown error",
			},
			500,
		);
	}
});

// Stream message
app.get("/stream", async (c) => {
	if (!currentSession) {
		return c.json({ success: false, error: "No active session" }, 400);
	}

	const url = new URL(c.req.url);
	const content = url.searchParams.get('content');

	if (!content) {
		return c.json({ success: false, error: "Content parameter is required" }, 400);
	}

	try {

		// Set up SSE headers
		c.header("Content-Type", "text/event-stream");
		c.header("Cache-Control", "no-cache");
		c.header("Connection", "keep-alive");

		const stream = new ReadableStream({
			async start(controller) {
				try {
					const messageStream = currentSession!.streamMessage(content);

					for await (const chunk of messageStream) {
						controller.enqueue(`data: ${JSON.stringify({ chunk })}\n\n`);
					}

					controller.enqueue(`data: ${JSON.stringify({ done: true })}\n\n`);
					controller.close();
				} catch (error) {
					controller.enqueue(
						`data: ${JSON.stringify({ error: error.message })}\n\n`,
					);
					controller.close();
				}
			},
		});

		return new Response(stream, {
			headers: {
				"Content-Type": "text/event-stream",
				"Cache-Control": "no-cache",
				Connection: "keep-alive",
			},
		});
	} catch (error) {
		return c.json(
			{
				success: false,
				error: error instanceof Error ? error.message : "Unknown error",
			},
			500,
		);
	}
});

// Execute tool
app.post("/tool", async (c) => {
	if (!currentSession) {
		return c.json({ success: false, error: "No active session" }, 400);
	}

	try {
		const { name, params } = await c.req.json();
		const execution = await currentSession.executeTool(name, params);
		return c.json({ success: true, execution });
	} catch (error) {
		return c.json(
			{
				success: false,
				error: error instanceof Error ? error.message : "Unknown error",
			},
			500,
		);
	}
});

// Get conversation
app.get("/conversation", async (c) => {
	if (!currentSession) {
		return c.json({ success: false, error: "No active session" }, 400);
	}

	const conversation = currentSession.getConversation();
	return c.json({ success: true, conversation });
});

// Get available tools
app.get("/tools", async (c) => {
	if (!currentSession) {
		return c.json({ success: false, error: "No active session" }, 400);
	}

	const tools = currentSession.getAvailableTools();
	return c.json({ success: true, tools });
});

// Get token usage
app.get("/tokens", async (c) => {
	if (!currentSession) {
		return c.json({ success: false, error: "No active session" }, 400);
	}

	const usage = currentSession.getTokenUsage();
	return c.json({ success: true, usage });
});

// Clear session
app.delete("/session", async (c) => {
	if (!currentSession) {
		return c.json({ success: false, error: "No active session" }, 400);
	}

	currentSession.clear();
	return c.json({ success: true });
});

const port = process.env.PORT ? parseInt(process.env.PORT) : 3000;

console.log(`🚀 Code Agent server starting on port ${port}`);

serve({
	fetch: app.fetch,
	port,
});

export { app };
