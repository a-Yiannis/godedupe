# Go Dedupe

A simple and efficient duplicate file finder written in Go.  
Most of this file (README.md) has been written by Gemini-2.5-Pro.

## How it works

1.  **Configuration**: The tool loads its configuration from a `config.json` file in the same directory. You can also provide a path to a configuration file as a command-line argument. If no `rootDirectory` directory is specified in the config, it will scan the current working directory.

2.  **Scanning**: It recursively scans the `rootDirectory` directory to find all files.

3.  **Grouping by Size**: It first groups files by their size, as files with different sizes cannot be duplicates.

4.  **Hashing**:
    *   **Partial Hash**: For files of the same size, it calculates a partial hash of the first few kilobytes (middle and end). This is a quick way to filter out non-duplicates.
    *   **Full Hash**: For files that have the same partial hash, it calculates a full hash of the entire file content. Files with the same full hash are considered duplicates.

5.  **Reporting**: It reports the sets of duplicate files found.

6.  **Recycling**: After reporting the duplicates, it will ask if you want to send the duplicate files (keeping the one with the oldest modification date) to the recycle bin. The recycled files are logged in `recycled.log`.

## Configuration

Configuration can be provided via a JSON file and/or command-line flags. Flags will override any values set in the configuration file.

### Configuration File

Create a `config.json` file (or use a different name and specify it with the `-config` flag) with the following structure:

```json
{
  "rootDirectory": "/path/to/your/folder",
  "directoriesToIgnore": [
    ".git",
    "node_modules"
  ],
  "extensionsToIgnore": [
    ".log",
    ".tmp"
  ],
  "pathsToIgnore": [
    "/path/to/your/folder/file.txt"
  ]
}
```

*   `rootDirectory`: The directory to scan for duplicates. If omitted, the current directory is used.
*   `directoriesToIgnore`: A list of directory names to exclude from the scan.
*   `extensionsToIgnore`: A list of file extensions to exclude from the scan.
*   `pathsToIgnore`: A list of full paths to files or directories to exclude from the scan.

### Command-line Flags

*   `-config <path>`: Path to the configuration file (default: `config.json`).
*   `-root <path>`: Root directory to scan (overrides config file).
*   `-ignore-dir <name>`: Directory to ignore (can be specified multiple times).
*   `-ignore-ext <ext>`: File extension to ignore (can be specified multiple times).
*   `-ignore-path <path>`: Path to ignore (can be specified multiple times).

**Example:**

```bash
go run . -root="C:\Users\YourUser\Documents" -ignore-ext=".tmp" -ignore-dir="cache"
```

## Disclaimer

The "recycle" feature will move files to your system's recycle bin. While this is generally safe, always make sure you have backups of important files before using this feature.
