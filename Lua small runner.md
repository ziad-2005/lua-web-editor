# Lua Runner --- Project Plan

## 1. Project Overview

**Lua Runner** is a web application that lets users write Lua code in a
browser, execute it on a backend, and inspect the result.

The experience should feel like a small online code-execution
environment:

1.  The user writes Lua code in an editor.
2.  The browser sends the source code to the backend.
3.  The backend creates an isolated execution job.
4.  A Lua runtime executes the code under strict resource and permission
    limits.
5.  The backend returns the output, errors, exit status, and execution
    metadata.
6.  The browser displays the result.

The project is both a useful tool and a learning project for
understanding backend engineering, process management, operating-system
isolation, and secure execution of untrusted code.

## 2. Goals

### Product goals

-   Provide a simple, responsive Lua editor in the browser.
-   Execute Lua code on the server.
-   Display standard output (`stdout`) and standard error (`stderr`)
    separately.
-   Report execution status, exit code, and duration.
-   Give each execution a clear lifecycle and prevent one user's job
    from affecting another user's job.

### Engineering goals

-   Keep the web application separate from the execution environment.
-   Make the execution mechanism replaceable: the first implementation
    may use containers, while a stronger isolation backend can be
    introduced later.
-   Enforce execution limits on the server, not merely in the Lua code.
-   Design the system so concurrent executions can be managed safely.
-   Keep the initial implementation small enough to understand and test.

## 3. Non-Goals for the First Version

The MVP will not attempt to be a full online IDE or a complete
replacement for ChatGPT's code-execution environment.

Out of scope initially:

-   User accounts and authentication.
-   Persistent user projects and cloud storage.
-   Collaboration and shared editing.
-   Multiple programming languages.
-   Package installation from the internet.
-   Long-running background applications or services.
-   Interactive terminal sessions and live `stdin`.
-   A distributed worker cluster.

These features can be considered after the basic execution path is
secure and reliable.

## 4. Proposed Technology Stack

  -----------------------------------------------------------------------
  Component               Proposed technology     Responsibility
  ----------------------- ----------------------- -----------------------
  Frontend                HTML, CSS, JavaScript   User interface

  Code editor             Monaco Editor           Lua source editing

  Backend                 Go                      HTTP API, validation,
                                                  job lifecycle

  Language runtime        Lua 5.4                 Execute submitted Lua
                                                  code

  Initial isolation       Dedicated container per Restrict filesystem,
                          execution               network, and resources

  Future isolation option gVisor or Firecracker   Stronger execution
                                                  isolation, depending on
                                                  deployment needs
  -----------------------------------------------------------------------

The stack is a proposal, not a requirement. Keep components replaceable
where practical.

## 5. High-Level Architecture

``` text
                 Browser
                    |
                    | POST /api/runs
                    | { "code": "print('Hello')" }
                    v
              Go HTTP Server
                    |
             Validate request
                    |
             Create execution job
                    |
                    v
             Execution Manager
                    |
       Apply timeout and resource limits
                    |
                    v
           Isolated Lua Runtime
             (one job at a time)
                    |
          Capture stdout / stderr
                    |
          Collect exit status
                    |
                    v
              Go HTTP Server
                    |
             JSON response
                    |
                    v
                 Browser
```

### Main components

#### Frontend

-   Code editor with Lua syntax highlighting.
-   Run button.
-   Output panel.
-   Error panel.
-   Execution status and duration.
-   Clear output and reset editor actions.

The frontend is not a security boundary. All validation and limits must
also be enforced by the backend.

#### Go HTTP server

-   Exposes the API.
-   Validates request size and input shape.
-   Creates and tracks execution jobs.
-   Passes source code to the execution manager.
-   Collects execution results.
-   Returns structured JSON responses.

#### Execution manager

-   Owns the lifecycle of each execution.
-   Starts the isolated runtime.
-   Applies time and resource limits.
-   Captures output.
-   Terminates jobs that exceed their limits.
-   Cleans up temporary files and execution resources.

The execution manager should be separated from HTTP handlers so that the
runtime implementation can be changed without redesigning the API.

#### Lua runtime

-   Runs a known, pinned Lua version.
-   Receives source code through a controlled input mechanism.
-   Has no access to the host's sensitive files.
-   Has no network access in the initial design.
-   Cannot exceed configured CPU, memory, process, or execution-time
    limits.

## 6. API Proposal

### `POST /api/runs`

Submit Lua code for execution.

Example request:

``` json
{
  "code": "print('Hello from Lua')"
}
```

Example response:

``` json
{
  "id": "run_123",
  "status": "completed",
  "stdout": "Hello from Lua\n",
  "stderr": "",
  "exit_code": 0,
  "duration_ms": 12
}
```

For the first MVP, the endpoint may wait for a short execution to finish
and return its result directly.

Later, execution can become asynchronous:

-   `POST /api/runs` creates a job and returns its ID.
-   `GET /api/runs/{id}` retrieves its status and result.
-   Optional cancellation endpoint terminates a running job.

The API should use explicit status values, for example:

-   `queued`
-   `running`
-   `completed`
-   `failed`
-   `timed_out`
-   `cancelled`

Do not expose host paths, internal command lines, environment secrets,
or raw infrastructure details in API responses.

## 7. Execution and Isolation Requirements

Running user-submitted code is a security-sensitive feature. Lua code
must be treated as untrusted.

Lua's standard libraries can expose operating-system and filesystem
functionality. Removing a Lua library or filtering source code is not
sufficient isolation.

The initial design should use a dedicated, disposable execution
environment for each job, with:

-   A non-root execution user.
-   A read-only runtime image where possible.
-   A temporary, job-specific writable directory.
-   No host filesystem mounts except narrowly scoped inputs and outputs.
-   Network access disabled by default.
-   CPU and memory limits.
-   A hard wall-clock timeout.
-   Process and file-descriptor limits.
-   Output-size limits to prevent memory or storage exhaustion.
-   Cleanup after completion, failure, or timeout.

Containers share the host kernel. They are useful for an initial
controlled deployment, but they are not automatically a strong security
boundary against hostile workloads. Do not expose a public service to
arbitrary code until its isolation and deployment have been reviewed and
tested.

For a stronger threat model, evaluate a sandbox such as gVisor or a
microVM-based design such as Firecracker. The choice depends on the host
environment, operational complexity, and required isolation.

## 8. Repository Structure Proposal

``` text
lua-runner/
├── cmd/
│   └── server/
│       └── main.go
├── internal/
│   ├── api/
│   │   ├── handlers.go
│   │   └── routes.go
│   ├── execution/
│   │   ├── manager.go
│   │   ├── runner.go
│   │   └── result.go
│   └── config/
│       └── config.go
├── web/
│   ├── index.html
│   ├── styles.css
│   └── app.js
├── runtime/
│   └── Dockerfile
├── tests/
├── go.mod
└── README.md
```

This is an initial organization proposal. Avoid creating abstractions
that do not yet solve a real problem, but keep HTTP handling separate
from execution and isolation logic.

## 9. Development Phases

### Phase 1 --- Local Lua execution

-   Install and pin Lua 5.4.
-   Write a small Go program that executes a Lua file.
-   Capture `stdout`, `stderr`, and exit status.
-   Add a timeout.
-   Test syntax errors, runtime errors, and infinite loops.

**Deliverable:** a working local Go-to-Lua execution prototype.

### Phase 2 --- HTTP API

-   Create the Go HTTP server.
-   Implement `POST /api/runs`.
-   Validate the request and enforce a maximum source-code size.
-   Return structured execution results.
-   Add API tests.

**Deliverable:** Lua code can be submitted and executed through HTTP.

### Phase 3 --- Browser interface

-   Add Monaco Editor.
-   Add a Run button.
-   Send code to the API.
-   Display output, errors, and execution status.
-   Handle loading and failed requests.

**Deliverable:** a usable browser-based Lua runner.

### Phase 4 --- Isolation and resource controls

-   Run jobs in disposable isolated environments.
-   Disable network access.
-   Restrict filesystem access.
-   Apply memory, CPU, process, output, and time limits.
-   Ensure cleanup on every exit path.
-   Test hostile and resource-exhausting programs.

**Deliverable:** a controlled execution service suitable for further
security review.

### Phase 5 --- Concurrency and job management

-   Introduce job IDs and explicit job states.
-   Add a bounded execution queue.
-   Limit concurrent jobs.
-   Add cancellation and timeout handling.
-   Prevent resource exhaustion through request and rate limits.

**Deliverable:** multiple users can submit jobs without uncontrolled
resource consumption.

### Phase 6 --- Future improvements

Potential additions:

-   Persistent projects and saved files.
-   Interactive input.
-   Execution history.
-   User accounts and quotas.
-   Multiple Lua versions.
-   Additional languages through a common runner interface.
-   Worker pools and distributed execution.

## 10. Important Design Decisions

Before implementation grows, document and agree on these decisions:

1.  **Execution model:** synchronous request/response initially, or
    asynchronous jobs from the beginning?
2.  **Isolation:** which container or sandbox runtime will be used, and
    what threat model does it address?
3.  **Deployment target:** local development machine, private server, or
    public service?
4.  **Persistence:** should files exist only for one execution, or
    survive between runs?
5.  **Concurrency:** how many executions may run simultaneously, and
    what are the per-user limits?
6.  **Lua environment:** standard Lua 5.4 libraries only, or a
    deliberately restricted set?
7.  **Input/output:** should the first version support only printed
    output, or also files and interactive input?

Do not treat these as cosmetic choices. They affect security, API
behavior, infrastructure cost, and the execution lifecycle.

## 11. Definition of Done for the MVP

The MVP is complete when a user can:

-   Open the web application.
-   Write valid Lua code.
-   Execute it through the Go backend.
-   See standard output and errors.
-   Receive a clear result for syntax errors and runtime failures.
-   Have infinite loops stopped by a server-enforced timeout.
-   Run code without gaining access to the host's sensitive files or
    network.
-   Use the service without one execution corrupting another execution's
    state.

## 12. Guiding Principle

**The backend manages execution; the sandbox contains execution.**

Keep the web server out of the business of directly running untrusted
code with its own permissions. Treat every submitted program as
potentially malicious, make resource usage bounded, and make cleanup
reliable.

Build the smallest complete system first, then strengthen its isolation
and scalability based on explicit requirements.
