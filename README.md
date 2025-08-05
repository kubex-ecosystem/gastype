
![Gastype Banner](./assets/top_banner.png)

---

**A flexible tool for parallel type checking of Go files, with dynamic CLI commands and extensibility for managing codebases efficiently.**

---

## **Table of Contents**
1. [About the Project](#about-the-project)
2. [Features](#features)
3. [Installation](#installation)
4. [Usage](#usage)
    - [CLI](#cli)
    - [Usage Examples](#usage-examples)
5. [Configuration](#configuration)
6. [Roadmap](#roadmap)
7. [Contributing](#contributing)
8. [Contact](#contact)

---

## **About the Project**
Gastype is a Go-based utility designed to analyze and validate codebases efficiently, with parallel type-checking capabilities. It provides developers with a streamlined way to ensure quality across Go projects, leveraging dynamic CLI commands and extensible functionalities.

**Why Gastype?**
- 🚀 **Parallel Execution**: Maximizes efficiency by processing multiple files at once.
- ⚙️ **Configurable**: Highly customizable for various workflows.
- 📂 **Clean and Modular**: Designed to be easy to maintain and extend.

---

## **Features**
⚡ **Parallel Type Checking**:
- Execute checks across multiple Go files simultaneously.
- Provides detailed feedback on errors and their locations.

💻 **Flexible CLI**:
- User-friendly CLI for managing type checks and file monitoring.
- Extensible commands for diverse use cases.

🔒 **Resilient and Safe**:
- Validates file structures before processing.
- Handles errors gracefully with detailed logging.

---

## **Installation**
Requirements:
- **Go** version 1.19 or later.

```bash
# Clone this repository
git clone https://github.com/rafa-mori/gastype.git

# Navigate to the project directory
cd gastype

# Build the binary using make
make build

# Install the binary using make
make install

# (Optional) Add the binary to the PATH to use it globally
export PATH=$PATH:$(pwd)
```

---

## **Usage**

### CLI
Here are some examples of commands you can execute with **gastype**:

```bash
# Perform type-checking in a directory
gastype check -d ./example -w 4 -o type_check_results.json

# Monitor Go files for changes and trigger type-checks
gastype watch -d ./example -w 4 -o type_check_results.json
```

---

### **Usage Examples**

#### **1. Check Go Files for Type Errors**

```bash
gastype check \
--dir ./example \
--workers 4 \
--output type_check_results.json
```

**Output:**

```json
{
  "package": "example_pkg",
  "status": "Success ✅",
  "error": ""
}
```

#### **2. Watch a Directory for File Changes**

```bash
gastype watch \
--dir ./example \
--workers 4 \
--output type_check_results.json \
--email gastype@gmail.com \
--token "secure-token"
```

**Output:**
```plaintext
Watching directory ./example...
File changes detected, type checking initiated.
```

---

### **Description of Commands and Flags**
- **`--dir`**: Specifies the directory containing Go files.
- **`--workers`**: Number of workers for parallel processing.
- **`--output`**: Output file for type-check results.
- **`--email`**: Email address for notifications.
- **`--token`**: Token for email notifications.

---

## **Configuration**
Gastype supports dynamic configurations via a JSON file, which allows users to set default values for directories, workers, and output files.

**Example Configuration**:
```json
{
  "dir": "./example",
  "workers": 4,
  "outputFile": "type_check_results.json",
  "email": "gastype@gmail.com",
  "token": "123456"
}
```

---

## **Roadmap**
🛠️ **Planned Enhancements**:
- Extend type-checking capabilities for advanced validation.
- Add support for custom file watchers and notifications.
- Improve the documentation and provide more practical examples.

---

## **Contributing**
Contributions are welcome! Feel free to open issues or submit pull requests. Check out the [Contributing Guide](CONTRIBUTING.md) for more details.

---

## **Contact**
💌 **Developer**:  
[Rafael Mori](mailto:faelmori@gmail.com)  
💼 [Follow me on GitHub](https://github.com/faelmori)  
I’m open to new collaborations and feedback. Don’t hesitate to reach out if you find this project interesting!
