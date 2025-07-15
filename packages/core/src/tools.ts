import { z } from "zod";

//  Simple Zod to JSON schema converter
function zodToJsonSchema(schema: z.ZodTypeAny): any {
  if (schema instanceof z.ZodObject) {
    const shape = schema.shape;
    const properties: Record<string, any> = {};
    const required: string[] = [];

    for (const [key, value] of Object.entries(shape)) {
      if (value instanceof z.ZodString) {
        properties[key] = { type: "string" };
        if (!value.isOptional()) {
          required.push(key);
        }
      } else if (value instanceof z.ZodOptional) {
        const inner = value.unwrap();
        if (inner instanceof z.ZodString) {
          properties[key] = { type: "string" };
        }
      } else if (value instanceof z.ZodDefault) {
        const inner = value.removeDefault();
        if (inner instanceof z.ZodString) {
          properties[key] = { type: "string" };
        }
      }
    }
    return {
      type: "object",
      properties,
      required,
    };
  }
  return { type: "object" };
}

export const ToolState = z.enum(["pending", "running", "completed", "error"]);
export type ToolState = z.infer<typeof ToolState>;

export const ToolExecution = z.object({
  id: z.string(),
  name: z.string(),
  state: ToolState,
  input: z.record(z.any()),
  output: z.any().optional(),
  error: z.string().optional(),
  startTime: z.number(),
  endTime: z.number().optional(),
});
export type ToolExecution = z.infer<typeof ToolExecution>;

export interface Tool {
  name: string;
  description: string;
  parameters: z.ZodSchema;
  execute: (params: any) => Promise<any>;
}

export interface GroqAITool {
  type: "function";
  function: {
    name: string;
    description: string;
    parameters: {
      type: "object";
      properties: Record<string, any>;
      required?: string[];
    };
  };
}

export class ToolExecutor {
  private tools = new Map<string, Tool>();
  private executions = new Map<string, ToolExecution>();

  registerTool(tool: Tool): void {
    this.tools.set(tool.name, tool);
  }

  async execute(name: string, params: any): Promise<ToolExecution> {
    const tool = this.tools.get(name);
    if (!tool) {
      throw new Error(`Tool ${name} not found`);
    }

    const execution: ToolExecution = {
      id: crypto.randomUUID(),
      name: tool.name,
      state: "pending",
      input: params,
      startTime: Date.now(),
    };

    this.executions.set(execution.id, execution);

    try {
      execution.state = "running";
      const validatedParams = tool.parameters.parse(params);
      const result = await tool.execute(validatedParams);

      execution.state = "completed";
      execution.output = result;
    } catch (error) {
      execution.state = "error";
      execution.error = error instanceof Error ? error.message : String(error);
    } finally {
      execution.endTime = Date.now();
    }

    return execution;
  }

  getExecution(id: string): ToolExecution | undefined {
    return this.executions.get(id);
  }

  getTools(): string[] {
    return Array.from(this.tools.keys());
  }

  getGroqAITools(): GroqAITool[] {
    return Array.from(this.tools.values()).map((tool) => ({
      type: "function",
      function: {
        name: tool.name,
        description: tool.description,
        parameters: zodToJsonSchema(tool.parameters),
      },
    }));
  }
}

// Built in tools
export const bashTool: Tool = {
  name: "bash",
  description: "Execute bash commands",
  parameters: z.object({
    command: z.string(),
  }),
  execute: async (params) => {
    const proc = Bun.spawn(["bash", "-c", params.command]);
    const output = await new Response(proc.stdout).text();
    const error = await new Response(proc.stderr).text();

    return {
      output: output.trim(),
      error: error.trim() || undefined,
      exitCode: proc.exitCode,
    };
  },
};

export const readFileTool: Tool = {
  name: "readFile",
  description: "Read a file from the filesystem",
  parameters: z.object({
    path: z.string(),
  }),
  execute: async (params) => {
    const file = Bun.file(params.path);
    const exists = await file.exists();

    if (!exists) {
      throw new Error(`File not found: ${params.path}`);
    }

    const content = await file.text();
    return {
      content,
      size: file.size,
    };
  },
};

export const writeFileTool: Tool = {
  name: "writeFile",
  description: "Write content to a file",
  parameters: z.object({
    path: z.string(),
    content: z.string(),
  }),
  execute: async (params) => {
    await Bun.write(params.path, params.content);
    return {
      path: params.path,
      size: params.content.length,
    };
  },
};

export const editFileTool: Tool = {
  name: "editFile",
  description: "Edit an existing file by replacing specific content",
  parameters: z.object({
    path: z.string(),
    oldContent: z.string(),
    newContent: z.string(),
  }),
  execute: async (params) => {
    const file = Bun.file(params.path);
    const exists = await file.exists();

    if (!exists) {
      throw new Error(`File not found: ${params.path}`);
    }

    const content = await file.text();

    if (!content.includes(params.oldContent)) {
      throw new Error(`Content not found in file: ${params.oldContent}`);
    }

    const newContent = content.replace(params.oldContent, params.newContent);
    await Bun.write(params.path, newContent);

    return {
      path: params.path,
      size: newContent.length,
    };
  },
};

export const makeDirTool: Tool = {
  name: "makeDir",
  description: "Create a directory (and parent directories if needed)",
  parameters: z.object({
    path: z.string(),
    recursive: z.boolean().default(true),
  }),
  execute: async (params) => {
    const proc = Bun.spawn(
      ["mkdir", params.recursive ? "-p" : "", params.path].filter(Boolean),
    );
    await proc.exited;

    if (proc.exitCode !== 0) {
      throw new Error(`Failed to create directory: ${params.path}`);
    }

    return {
      path: params.path,
      created: true,
    };
  },
};

export const listFilesTool: Tool = {
  name: "list_files",
  description: "List files in a directory",
  parameters: z.object({
    path: z.string().default("."),
  }),
  execute: async (params) => {
    const proc = Bun.spawn(["ls", "-la", params.path]);
    const output = await new Response(proc.stdout).text();

    return {
      files: output.split("\n").filter((line) => line.trim()),
      path: params.path,
    };
  },
};

if (import.meta.main) {
  const executor = new ToolExecutor();
  executor.registerTool(bashTool);
  executor.registerTool(listFilesTool);

  // Test bash tool
  console.log("Testing bash tool...");
  const bashResult = await executor.execute("bash", {
    command: 'echo "Hello World"',
  });
  console.log("Bash result:", bashResult);

  // Test list files
  console.log("Testing list_files tool...");
  const listResult = await executor.execute("list_files", { path: "." });
  console.log("List result:", listResult);
}
