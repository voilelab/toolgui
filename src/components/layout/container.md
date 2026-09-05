# Container

Container is the most basic layout component.
Its definition is in tgframe.
The Main "container" and Sidebar "container" are Containers.

## Usage

To create a container under a container. Just call the function:

```go
func (c *Container) AddContainer(id string) *Container
```

* ID should be unique.
