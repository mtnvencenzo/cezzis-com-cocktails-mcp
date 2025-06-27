# template
Template repo for creating other repos from

Building the app for local usage.

Navigate to the /src diectory and run:
``` bash
make forclaude
```

# Setting up with claude desktop
Make sure an entry exists in this file for the deploy path of the local exe
> C:\Users\rvecc\AppData\Roaming\Claude\claude_desktop_config.json

``` json
{
  "mcpServers": {
	"mcp-cocktails-go": {
      "command": "D:\\Github\\cezzis-com-cocktails-mcp\\dist\\cezzis-cocktails.exe"
    }
  }
}

After changing the file, it's not enough to close and reopen Claude Desktop as it runs in the background.  You must open Claude and exit out of it using the file menu.  Then repoen it.  This ensures the settings are re-loaded.
```

## Helpful Links
https://github.com/AzureAD/microsoft-authentication-library-for-go