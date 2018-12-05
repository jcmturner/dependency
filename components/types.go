package components

type Type string
type Class string

const (
	TypeJava       Type = "java"
	TypeDotNet     Type = "dotnet"
	TypeJavaScript Type = "javascript"
	TypePython     Type = "python"
	TypeOSNative   Type = "os-native"

	ClassLib     Class = "library"
	ClassRuntime Class = "runtime"
	ClassOS      Class = "os"
)
