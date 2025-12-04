<div align="center">
  <h1>Agentic Workflow: The CodeAct Loop</h1>
  <p>This diagram illustrates the core execution loop of the Construct agent system.</p>
</div>

```mermaid
graph TD
    %% Nodes
    Start([Start Task]) --> AwaitInput
    
    subgraph "Task Reconciler Loop"
        direction TB
        AwaitInput["<b>Await Input</b><br/>(User Message / Event)"]
        
        ComputeStatus{Check Status}
        
        InvokeModel["<b>Invoke Model</b><br/>(LLM Thinking)"]
        
        DecideAction{Action?}
        
        ExecuteTools["<b>Execute Tools</b><br/>(CodeAct Interpreter)"]
        
        subgraph "Tool Execution (Sandboxed VM)"
            Script[Run JS Script]
            ToolCall1[Call: execute_command]
            ToolCall2[Call: edit_file]
            ToolCall3[Call: handoff]
        end
        
        Persist["<b>Persist Results</b><br/>(Save to Memory)"]
    end

    %% Edges
    AwaitInput --> ComputeStatus
    ComputeStatus -- New Message --> InvokeModel
    
    InvokeModel -- "Generates Text" --> AwaitInput
    InvokeModel -- "Generates Code" --> ExecuteTools
    
    ExecuteTools --> Script
    Script -- "Sequential" --> ToolCall1
    Script -- "Sequential" --> ToolCall2
    Script -- "Sequential" --> ToolCall3
    
    ToolCall1 --> Persist
    ToolCall2 --> Persist
    ToolCall3 --> Persist
    
    Persist -- "Tool Outputs" --> InvokeModel
    
    %% Styling
    classDef default fill:#f9f9f9,stroke:#333,stroke-width:2px;
    classDef decision fill:#e1f5fe,stroke:#01579b,stroke-width:2px;
    classDef process fill:#e8f5e9,stroke:#2e7d32,stroke-width:2px;
    classDef loop fill:#fff3e0,stroke:#ef6c00,stroke-width:2px,stroke-dasharray: 5 5;
    
    class ComputeStatus,DecideAction decision;
    class InvokeModel,ExecuteTools,Persist process;
    class ToolCall1,ToolCall2,ToolCall3 loop;
```

### Workflow Explanation

1.  **Sequential Loop**: The primary architecture is a **sequential loop**. The agent thinks, acts, observes the result, and thinks again.
2.  **Parallelism**: While the *loop* is sequential, the **CodeAct** paradigm allows the agent to write a script that performs multiple actions in a batch.
    *   *Example*: The agent can write a script that edits 3 different files. The Interpreter runs this script. While the script runs sequentially line-by-line, from the "Thinking" perspective, it's a single "Act" step that accomplishes multiple things before the LLM needs to think again.
3.  **Recursion/Looping**: The system loops automatically.
    *   If the Agent runs a command and gets an error, the `Persist` step saves that error.
    *   The flow goes back to `Invoke Model`.
    *   The LLM sees the error and generates a *new* script to fix it.
    *   This continues until the task is complete (`submit_report`) or the agent needs human help (`ask_user`).
