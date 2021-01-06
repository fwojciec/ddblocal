package mocks

//go:generate moq -out string_generator.go -pkg mocks .. StringGenerator
//go:generate moq -out presence_checker.go -pkg mocks .. PresenceChecker
//go:generate moq -out executor_terminator.go -pkg mocks .. ExecutorTerminator
//go:generate moq -out client_initializer.go -pkg mocks .. ClientInitializer
