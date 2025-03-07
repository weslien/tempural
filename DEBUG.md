# Debugging Tempural

This document provides guidance on debugging and profiling Tempural, including specific instructions for using Cursor IDE.

## Debugging in Cursor IDE

Cursor IDE makes it easy to debug Go applications. Here's how to set up debugging for Tempural:

1. **Enable Debugging in Cursor**:
   - Click on the Debug icon in the sidebar (or press `Cmd+Shift+D` on macOS, `Ctrl+Shift+D` on Windows/Linux)
   - Click on "create a launch.json file" if you don't already have one

2. **Configure launch.json**:
   Add the following configuration to your `.vscode/launch.json` file:
   ```json
   {
     "version": "0.2.0",
     "configurations": [
       {
         "name": "Debug Tempural",
         "type": "go",
         "request": "launch",
         "mode": "auto",
         "program": "${workspaceFolder}/main.go",
         "args": ["list"],
         "cwd": "${workspaceFolder}"
       },
       {
         "name": "Debug Tempural Start",
         "type": "go",
         "request": "launch",
         "mode": "auto",
         "program": "${workspaceFolder}/main.go",
         "args": ["start", "-t", "YourWorkflowType", "-i", "{\"key\": \"value\"}"],
         "cwd": "${workspaceFolder}"
       }
     ]
   }
   ```

3. **Set Breakpoints**:
   - Open any Go file in the project
   - Click in the gutter next to a line number to set a breakpoint
   - You can also set conditional breakpoints by right-clicking on a breakpoint

4. **Start Debugging**:
   - Select the appropriate configuration from the dropdown in the Debug panel
   - Click the green play button or press F5 to start debugging
   - The program will pause when it hits your breakpoints

## Using Profiling Features

Tempural includes built-in profiling features that can be used to analyze performance and diagnose issues.

### CPU Profiling

To capture a CPU profile:

```bash
./tempural --cpu-profile=cpu.prof list
```

This will save a CPU profile to `cpu.prof` which you can analyze with:

```bash
go tool pprof cpu.prof
```

Inside the pprof tool, you can use commands like:
- `top` - Show top functions by CPU usage
- `web` - Launch a web browser to visualize the profile (requires graphviz)
- `list functionName` - Show source code with profiling information

### Memory Profiling

To capture a memory profile:

```bash
./tempural --mem-profile=mem.prof list
```

This will save a memory profile to `mem.prof` which you can analyze with:

```bash
go tool pprof mem.prof
```

### Runtime Profiling Server

For continuous profiling during development:

```bash
./tempural --pprof list
```

This starts a profiling server on port 6060 (configurable with `--pprof-port`). Open your browser to:
- http://localhost:6060/debug/pprof/ - Index page with various profile types
- http://localhost:6060/debug/pprof/heap - Memory allocation profile
- http://localhost:6060/debug/pprof/goroutine - Current goroutines
- http://localhost:6060/debug/pprof/profile?seconds=30 - 30-second CPU profile

### Generating Visualizations

With graphviz installed, you can generate visual representations of profiles:

```bash
go tool pprof -png -output=cpu.png cpu.prof
go tool pprof -svg -output=mem.svg mem.prof
```

## Tips for Effective Debugging

1. **Use Debug Logging**:
   ```bash
   ./tempural --debug command [args...]
   ```

2. **Combine Techniques**:
   ```bash
   # Debug AND profile
   ./tempural --debug --cpu-profile=cpu.prof list
   ```

3. **Profile in Production**:
   For production use, you can enable the profiling server on a protected port to diagnose issues in running services.

4. **Unit Test Debugging**:
   You can also debug unit tests in Cursor IDE by selecting the "Debug Test" option that appears above each test function. 