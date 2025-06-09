# Golang Auth API - Hexagonal Architecture

This project is a Golang Auth API that follows the hexagonal (also known as Ports and Adapters) architecture pattern. It provides a flexible and modular structure that separates the concerns of the application into different layers, making it easier to maintain, extend, and test.

### Top-level Files

- `Dockerfile`: Contains instructions for building a Docker container for the application.
- `Taskfile.yml`: Contains the list of tasks and their configurations for the Task build tool.
- `main.go`: The main entry point for the application.

### Internal Directory

The `internal` directory contains the main components of the hexagonal architecture:

#### App Layer

The `app` directory contains the application's core business logic and is organized as follows:

- `github.go`, `gitlab.go`, `google.go`: Implementations of authentication providers.
- `gqlgen.yml`: Configuration file for the Golang GraphQL library, gqlgen.
- `graph`: Contains GraphQL-related files such as resolvers, schema, and generated code.
- `grpc-server.go`: Implements a gRPC server for the application.
- `main.go`: The main entry point for the app layer.

#### Domain Layer

The `domain` directory contains the domain model and interfaces that define the contract for communication between the app and framework layers:

- `domain.go`: Contains the domain model and interfaces for the application.
- `entities.go`: Defines the main domain entities.
- `impl.go`: Provides implementations of the domain interfaces.
- `main.go`: The main entry point for the domain layer.
- `ports.go`: Defines the ports (interfaces) that are used to interact with external components.

#### Framework Layer

The `framework` directory contains the infrastructure and external dependencies required by the application:

- `main.go`: The main entry point for the framework layer, which sets up and configures the necessary infrastructure components.

## Usage

To build and run the project, navigate to the root directory and execute the following commands:

```
go build
./auth-api
```

## Contributing

To contribute to the project, follow these steps:

1. Fork the repository and clone it to your local machine.
2. Create a new branch with a descriptive name, like `feature/add-auth-provider`.
3. Make your changes, following the existing project structure and adhering to the hexagonal architecture pattern.
4. Ensure that your code is formatted according to the project's coding standards and write tests if necessary.
5. Commit your changes and push them to your forked repository.
6. Open a pull request against the original repository with a clear description of your changes and why they are necessary.

Please follow the project's code of conduct and best practices while contributing.

## License

This project is licensed under the [MIT License](../LICENSE). All code in the repository is subject to this license unless stated otherwise.

## Contact

If you have any questions, issues, or suggestions regarding the project, please open an issue on the project's GitHub repository or contact the project maintainers via email.

Thank you for using and contributing to our Golang Auth API!
