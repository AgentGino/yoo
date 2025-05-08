import sys
import urllib.request
import json
import os
from pathlib import Path

CONFIG_DIR = Path.home() / ".yo"
CONFIG_FILE = CONFIG_DIR / "config.json"
DEFAULT_CONFIG = {
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

def ensure_config_exists():
    """Ensures the configuration file and directory exist, creating them with defaults if necessary."""
    if not CONFIG_FILE.exists():
        CONFIG_DIR.mkdir(parents=True, exist_ok=True)
        with open(CONFIG_FILE, 'w') as f:
            json.dump(DEFAULT_CONFIG, f, indent=2)
        print(f"Configuration file created at: {CONFIG_FILE}", file=sys.stderr)
        print("Please update it with your OpenAI API key.", file=sys.stderr)

def load_config():
    """Loads the configuration from the JSON file."""
    ensure_config_exists() # Ensure config is there before loading
    try:
        with open(CONFIG_FILE, 'r') as f:
            return json.load(f)
    except (json.JSONDecodeError, FileNotFoundError) as e:
        print(f"Error loading configuration: {e}", file=sys.stderr)
        sys.exit(1)

def call_openai_llm(input_text, mode, config):
    """
    Call OpenAI's LLM with the given input text and mode, using settings from the config.

    Args:
        input_text (str): The user's query or input.
        mode (str): The mode of operation (e.g., "terminal", "code"),
                    which determines the system prompt and model.
        config (dict): The loaded configuration dictionary.

    Returns:
        str: The content of the LLM's response, stripped of leading/trailing whitespace.
             Exits the program with an error message if any step fails.
    """
    ai_config = config.get("ai", {})
    # Prioritize API key from environment variable (OPENAI_API_KEY), then from config file.
    api_key = os.environ.get('OPENAI_API_KEY') or ai_config.get("apikey")
    # Default to OpenAI chat completions URL if not specified in config.
    api_url = ai_config.get("apiurl", "https://api.openai.com/v1/chat/completions")
    
    # Check if API key is missing or is the placeholder value.
    if not api_key or api_key == "YOUR_OPENAI_API_KEY":
        print("Error: OpenAI API key not found in environment variables (OPENAI_API_KEY) or config file.", file=sys.stderr)
        print(f"Please set it in your environment or in {CONFIG_FILE}", file=sys.stderr)
        sys.exit(1)

    prompts_config = config.get("prompts", {})
    mode_config = prompts_config.get(mode)

    # Ensure the selected mode is defined in the configuration.
    if not mode_config:
        print(f"Error: Mode '{mode}' not found in the configuration file.", file=sys.stderr)
        if prompts_config and isinstance(prompts_config, dict): # Check if prompts_config is a non-empty dict
            available_modes = ', '.join(prompts_config.keys())
            if available_modes:
                print(f"Available modes are: {available_modes}", file=sys.stderr)
            else:
                print("No modes are defined in the configuration.", file=sys.stderr)
        else:
            print("Prompts configuration is missing or invalid.", file=sys.stderr)
        sys.exit(1)

    system_prompt = mode_config.get("prompt")
    # Use model specified for the mode; fallback to default AI model from config, then to a hardcoded default ("gpt-4o").
    model = mode_config.get("model") or ai_config.get("defaultmodel", "gpt-4o")

    # Ensure a system prompt is defined for the selected mode.
    if not system_prompt:
        print(f"Error: Prompt for mode '{mode}' not found in the configuration.", file=sys.stderr)
        sys.exit(1)
    
    # Construct the message payload for the OpenAI API.
    system_message = {
        "role": "system",
        "content": system_prompt
    }
    user_message = {"role": "user", "content": input_text}
    messages = [system_message, user_message]
    
    body = {
        "model": model,
        "messages": messages,
        "temperature": 0  # Set temperature to 0 for more deterministic and factual output.
    }
    data = json.dumps(body).encode('utf-8') # Prepare the request body as JSON.
    
    headers = {
        "Content-Type": "application/json",
        "Authorization": f"Bearer {api_key}"
    }
    
    req = urllib.request.Request(api_url, data=data, headers=headers, method="POST")
    
    # Initialize variables for error reporting clarity
    response_data_str = "Response data not captured before error."
    response_json = None

    try:
        with urllib.request.urlopen(req) as response:
            response_bytes = response.read()
            response_data_str = response_bytes.decode('utf-8') # For logging in case of JSON error
            response_json = json.loads(response_data_str)
            
            # Extract the content from the LLM's response with robust validation.
            if (isinstance(response_json, dict) and
                "choices" in response_json and
                isinstance(response_json["choices"], list) and
                len(response_json["choices"]) > 0 and
                isinstance(response_json["choices"][0], dict) and
                "message" in response_json["choices"][0] and
                isinstance(response_json["choices"][0]["message"], dict) and
                "content" in response_json["choices"][0]["message"] and
                isinstance(response_json["choices"][0]["message"]["content"], str)):
                content = response_json["choices"][0]["message"]["content"]
                return content.strip()
            else:
                print("Error: Unexpected LLM response structure. Missing 'choices[0].message.content' or type mismatch.", file=sys.stderr)
                print(f"Full Response JSON: {json.dumps(response_json, indent=2) if response_json else response_data_str}", file=sys.stderr)
                sys.exit(1)

    except urllib.error.HTTPError as e:  # Handle HTTP errors from the API (e.g., 4xx, 5xx).
        error_body_str = "Could not read error body."
        try:
            error_body_bytes = e.read() # e.read() returns bytes
            error_body_str = error_body_bytes.decode('utf-8')
        except Exception as read_decode_err:
            error_body_str = f"(Could not read or decode error body: {read_decode_err})"
        print(f"Error: HTTP error: {e.code} {e.reason}", file=sys.stderr)
        print(f"Response body: {error_body_str}", file=sys.stderr)
        sys.exit(1)
    except urllib.error.URLError as e:  # Handle other network errors (e.g., DNS failure, connection refused).
        print(f"Error: Network error while contacting LLM: {e.reason}", file=sys.stderr)
        sys.exit(1)
    except json.JSONDecodeError as e:  # Handle errors in parsing JSON response.
        print(f"Error: Invalid JSON response from LLM: {e}", file=sys.stderr)
        # response_data_str contains the raw string that failed to parse
        print(f"Raw response data: {response_data_str}", file=sys.stderr)
        sys.exit(1)
    # KeyError should be caught by the structural validation above, but kept as a safeguard.
    except KeyError as e: 
        print(f"Error: Missing key '{e}' in LLM response.", file=sys.stderr)
        # If json.loads succeeded, response_json has the dict, otherwise log raw string.
        logged_data = json.dumps(response_json, indent=2) if response_json else response_data_str
        print(f"Response data: {logged_data}", file=sys.stderr)
        sys.exit(1)
    except Exception as e:  # Catch any other unexpected errors during API call or response processing.
        print(f"An unexpected error occurred: {e}", file=sys.stderr)
        # For debugging, you might want to log the traceback here.
        # import traceback; print(traceback.format_exc(), file=sys.stderr)
        sys.exit(1)

def main():
    """
    Main function to run the Yo CLI tool.

    This function orchestrates the command-line interface of the Yo tool.
    It performs the following steps:
    1. Ensures that the necessary configuration directory and file exist, creating
       them with default values if they are missing (`ensure_config_exists`).
    2. Loads the configuration, which includes API settings and predefined prompts
       for different modes (`load_config`).
    3. Parses command-line arguments to get the desired mode (e.g., "terminal", "code")
       and the user's query.
    4. Validates that the correct number of arguments are provided. If not, it prints
       usage instructions, lists available modes (if configured), and exits.
    5. Calls the `call_openai_llm` function with the user's query, selected mode, and
       loaded configuration. This function handles the interaction with the OpenAI API.
    6. Prints the response from the LLM to standard output. This response is typically
       a shell command or a piece of code, and it's formatted with a leading indent.
    7. Handles errors at various stages (e.g., configuration issues, invalid arguments,
       API call failures). Informative messages are printed to standard error, and
       the program exits with a non-zero status code.
    """
    ensure_config_exists() # Step 1: Ensure configuration is in place.
    config = load_config() # Step 2: Load configuration.

    # Step 3 & 4: Parse and validate command-line arguments.
    if len(sys.argv) < 3:
        print(f"Usage: {sys.argv[0]} <mode> <your query>", file=sys.stderr)
        print("Example: yo terminal list all files in the current directory", file=sys.stderr)
        prompts_cfg = config.get("prompts")
        if prompts_cfg and isinstance(prompts_cfg, dict):
            available_modes = ', '.join(prompts_cfg.keys())
            if available_modes: # Ensure there are modes to print
                 print(f"Available modes: {available_modes}", file=sys.stderr)
            else:
                print("No modes defined in configuration.", file=sys.stderr)
        else:
            print("Prompts configuration is missing or not found.", file=sys.stderr)
        sys.exit(1) # Exit if arguments are insufficient.
    
    mode = sys.argv[1]  # The first argument after the script name is the mode.
    input_text = ' '.join(sys.argv[2:])  # The rest of the arguments form the input query.
    
    # Step 5: Call the LLM.
    # call_openai_llm is designed to exit on any error, so if it returns,
    # 'command' should contain a valid string response from the LLM.
    command = call_openai_llm(input_text, mode, config)
    
    # Step 6 & 7: Print the response or handle the case of an empty (but valid) response.
    if command:  # Check if the returned command is not an empty string.
        print(f"   {command}")  # Print the command with leading spaces for consistent output.
    else:
        # This case is hit if the LLM returns an empty string (after stripping whitespace).
        # call_openai_llm exits on API/network errors, but an empty content from a 200 OK is possible.
        print("Error: Received an empty response from the LLM.", file=sys.stderr)
        sys.exit(1)

if __name__ == "__main__":
    main()