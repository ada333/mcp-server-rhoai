# mcp-server-RHOAI
This project aims to implement MCP servers working with RHOAI and is used as a bachelors thesis.

# Setting up MCP server in Cursor
- Compile the main.go function (in repo -> go build -o any_name)
- In Cursor press Ctrl + Shift + P and type in Open MCP Settings
- Click on New MCP server
Example code you can add to make new MCP server (the command is path to the binary you build):
```
{
  "mcpServers": {
    "greeter": {
      "command": "/home/amaly/MCP-test/greeter"
    },
    "podslist": {
      "command": "/home/amaly/MCP-test/podslist"
    },
    "workbenchlist": {
      "command": "/home/amaly/MCP-test/WorkbenchList"
    }
  }
}
```
- Check that server is enabled in Cursor MCP settings and has some tools you can use
- now you can use the mcp tools just by talking to AI agent in Cursor (example prompt: find me workbenches in namespace mcp-test)


(to use the tools operating with OpenShift cluster you need to be logged in)

