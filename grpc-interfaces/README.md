# gRPC Interfaces Folder - Golang Project

Welcome to the gRPC Interfaces folder of our Golang project. This folder contains the Protocol Buffer definitions (`.proto` files) and the auto-generated gRPC code for the various services and messages used in our project. It provides an organized structure for managing gRPC communication between different components within the project.

## Structure

The structure of the gRPC Interfaces folder is as follows:

```
grpc/
  ├── interfaces/
  │   ├── service1/
  │   │   ├── service1.proto
  │   │   ├── service1.pb.go
  │   │   └── service1_grpc.pb.go
  │   ├── service2/
  │   │   ├── service2.proto
  │   │   ├── service2.pb.go
  │   │   └── service2_grpc.pb.go
  │   └── ...
  ├── Makefile
  └── ...
```

Each subdirectory within `grpc/interfaces` represents a separate gRPC service. Inside each service folder, you will find the `.proto` file containing the service and message definitions, as well as the auto-generated `.pb.go` and `_grpc.pb.go` files containing the corresponding Golang code.

## Usage

To use a gRPC service from the `grpc/interfaces` folder, simply import it in your Go code using the project's import path. For example, if your project's import path is `github.com/username/project`, and you want to import `service1`, you would do:

```go
import "github.com/username/project/grpc/interfaces/service1"
```

You can then use the types and functions provided by the service to implement gRPC communication between different components of your project.

## Generating gRPC Code

To generate the Golang code for the gRPC services, you will need to have the Protocol Buffer Compiler (`protoc`) and the gRPC-Go plugin installed on your system. Once you have them installed, you can use the provided `Makefile` to generate the necessary files:

```
make generate
```

This command will run `protoc` with the gRPC-Go plugin and generate the `.pb.go` and `_grpc.pb.go` files for each service in the `grpc/interfaces` folder. If you add, update, or remove a `.proto` file, make sure to run the `make generate` command to keep the generated code in sync.

## Contributing

To contribute to the gRPC Interfaces folder:

1. Fork the repository and clone it to your local machine.
2. Create a new branch with a descriptive name, like `feature/service1-update`.
3. Make your changes or add new `.proto` files to the `grpc/interfaces` folder.
4. Run `make generate` to update the auto-generated Golang code.
5. Commit your changes and push them to your forked repository.
6. Open a pull request against the original repository with a clear description of your changes and why they are necessary.

Please follow the project's code of conduct and best practices while contributing.

## License

This project is licensed under the [MIT License](../LICENSE). All gRPC Interfaces in the `grpc/interfaces` folder are subject to this license unless stated otherwise.

## Contact

If you have any questions, issues, or suggestions regarding the gRPC Interfaces folder, please open an issue on the project's GitHub repository or contact the project maintainers via email.

Thank you for using and contributing to our Golang project!
