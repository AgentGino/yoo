# Yo CLI Tool

`yo` is a command-line interface (CLI) tool that allows you to interact with OpenAI's Large Language Models (LLMs) directly from your terminal. You can use it to get shell commands, code snippets, or general assistance.

## Features

-   Interact with different OpenAI models (configurable).
-   Define custom prompts and modes for various tasks (e.g., generating terminal commands, code).
-   Securely stores your OpenAI API key.

## Prerequisites

-   Python 3.x
-   An OpenAI API key.

## Installation & Configuration

1.  **Clone the repository or download `yo.py`**.
2.  **Set up your OpenAI API Key**:
    *   The script will look for the `OPENAI_API_KEY` environment variable first.
    *   Alternatively, you can place your API key in the configuration file.
3.  **First Run & Configuration File**:
    When you run `yo.py` for the first time, it will automatically create a configuration directory and file at `~/.yo/config.json` with default settings.

    ```json
    {
      "ai": {
        "apikey": "YOUR_OPENAI_API_KEY",
        "apiurl": "https://api.openai.com/v1/chat/completions",
        "defaultmodel": "gpt-4o"
      },
      "prompts": {
        "terminal": {
          "prompt": "You are a helpful assistant that provides mac os shell commands based on user requests. Respond with ONLY the raw shell command, ensuring it is POSIX-compliant. Do not include explanations or markdown formatting.",
          "model": "gpt-3.5-turbo"
        },
        "code": {
          "prompt": "You are a helpful coding assistant. Provide concise code snippets or explanations.",
          "model": "gpt-4o"
        }
      }
    }
    ```
    *   **Important**: You **must** update `YOUR_OPENAI_API_KEY` in this file with your actual OpenAI API key if it's not set as an environment variable.
    *   You can customize the `apiurl`, `defaultmodel`, and the prompts for different modes.

## Usage

The basic syntax is:

```bash
python yo.py <mode> <your query>
```

Or, if you make `yo.py` executable and place it in your PATH (e.g., as `yo`):

```bash
yo terminal list all .py files
```

### Making `yo.py` Executable and Easily Accessible (Optional)

For more convenient use, you can make the `yo.py` script directly runnable. Here's how for different operating systems:

**macOS & Linux:**

There are two main approaches:

1.  **Make the script executable and place it in a directory in your PATH:**
    *   Open your terminal.
    *   Navigate to the directory where `yo.py` is located.
    *   Make the script executable:
        ```bash
        chmod +x yo.py
        ```
    *   Choose a directory from your PATH. Common choices include `/usr/local/bin` (for system-wide access, often requires `sudo`) or `~/.local/bin` (for user-specific access; you might need to create this directory and add it to your PATH if it doesn't exist).
        To check your PATH:
        ```bash
        echo $PATH
        ```
        If `~/.local/bin` is not in your PATH, add it to your shell configuration file (`~/.zshrc` for Zsh, `~/.bashrc` for Bash, or `~/.profile`):
        ```bash
        export PATH="$HOME/.local/bin:$PATH"
        ```
        Then, source the file (e.g., `source ~/.zshrc`).
    *   Move `yo.py` to your chosen directory, optionally renaming it to `yo` for brevity:
        ```bash
        # Example for /usr/local/bin (might require sudo)
        sudo mv yo.py /usr/local/bin/yo
        
        # Example for ~/.local/bin
        mkdir -p ~/.local/bin # Create if it doesn't exist
        mv yo.py ~/.local/bin/yo
        ```
    *   Now you should be able to run it directly:
        ```bash
        yo terminal list files
        ```

2.  **Create an alias:**
    *   Open your shell configuration file (e.g., `~/.zshrc` for Zsh, `~/.bashrc` for Bash).
    *   Add the following line, replacing `/path/to/your/yo.py` with the actual absolute path to your `yo.py` script:
        ```bash
        alias yo='python3 /path/to/your/yo.py'
        ```
        Or, if you've already made `yo.py` executable and prefer to run it directly (without `python3` prefix):
        ```bash
        alias yo='/path/to/your/yo.py' 
        ```
    *   Save the file (e.g., `Ctrl+X` then `Y` then `Enter` in nano).
    *   Reload your shell configuration:
        ```bash
        # For Zsh
        source ~/.zshrc
        # For Bash
        source ~/.bashrc
        ```
    *   Now you can use the `yo` alias:
        ```bash
        yo terminal list files
        ```

**Windows:**

1.  **Ensure Python is in your PATH:**
    *   When installing Python, make sure to check the box that says "Add Python to PATH".
    *   If Python is installed and in your PATH, you can run Python scripts from any directory using `python yo.py` or `py yo.py`.

2.  **Add the script's directory to the PATH environment variable:**
    *   Find the full path to the directory where `yo.py` is located (e.g., `C:\Users\YourUser\Scripts`).
    *   Search for "environment variables" in the Windows search bar and select "Edit the system environment variables".
    *   In the System Properties window, click the "Environment Variables..." button.
    *   Under "User variables" (or "System variables" for all users), find the variable named `Path` and select it.
    *   Click "Edit...".
    *   Click "New" and add the full path to the directory containing `yo.py`.
    *   Click "OK" on all open windows to save the changes.
    *   You may need to restart your Command Prompt or PowerShell for the changes to take effect.
    *   Now, you can navigate to any directory in your terminal and run `yo.py` (or `yo` if you rename it to `yo.py` to `yo.py` in that directory):
        ```powershell
        # Assuming you renamed yo.py to yo.py or have yo.py in the directory
        python yo.py terminal list files 
        # or just yo.py if .PY is in PATHEXT
        yo.py terminal list files
        ```

3.  **Create a `.bat` file for a custom command (like an alias):**
    *   Open Notepad or any text editor.
    *   Add the following line, replacing `C:\path\to\your\yo.py` with the actual full path to your `yo.py` script:
        ```batch
        @echo off
        python C:\path\to\your\yo.py %*
        ```
        The `%*` passes all command-line arguments to the script.
    *   Save this file as `yo.bat` in a directory that is already in your PATH (e.g., `C:\Windows\System32`, though saving to a user-specific scripts directory and adding that to PATH is often preferred).
    *   Now you can open Command Prompt or PowerShell and type:
        ```powershell
        yo terminal list files
        ```

If you add `#!/usr/bin/env python3` (or the correct path to your python interpreter) as the first line of `yo.py` on macOS/Linux, after `chmod +x yo.py`, you can run it as `./yo.py <mode> <query>` directly if it's in the current directory, or `yo <mode> <query>` if it's in a PATH directory and renamed.

### Available Modes

The default modes defined in `config.json` are:

*   `terminal`: Generates shell commands based on your query.
*   `code`: Provides code snippets or explanations.

You can see the available modes by running the script with insufficient arguments:
```bash
python yo.py
```
or
```bash
yo
```

### Examples

*   **Get a shell command to list all files in the current directory:**
    ```bash
    python yo.py terminal list all files in the current directory
    ```
    Or if `yo.py` is executable and in PATH as `yo`:
    ```bash
    yo terminal list all files in the current directory
    ```

*   **Ask for a Python code snippet to read a file:**
    ```bash
    python yo.py code python read file
    ```
    Or if `yo.py` is executable and in PATH as `yo`:
    ```bash
    yo code python read file
    ```

## How it Works

The script sends your query, along with a system prompt defined for the chosen mode, to the specified OpenAI model via the Chat Completions API. It then prints the model's response to the standard output.

## Customization

You can customize the behavior of `yo` by editing the `~/.yo/config.json` file:

*   **Add new modes**: Define new entries under the `"prompts"` section with a unique mode name, a system `prompt`, and optionally a specific `model`.
*   **Change default model**: Update the `"defaultmodel"` under the `"ai"` section.
*   **Change API endpoint**: Modify the `"apiurl"` if you are using a proxy or a different OpenAI-compatible API.

## Contributing

Feel free to open issues or submit pull requests if you have suggestions or improvements! 

## License
MIT