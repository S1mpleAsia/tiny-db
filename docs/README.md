## TinyDB
A lightweight, small and simple DB written in Go. The system is intended for pedagogical use only. The project is inspried from the book [*Database Design and Implementation](https://link.springer.com/book/10.1007/978-3-030-33836-7)

### Architecture
TinyDB is built using a bottom-up modular approach consisting of:
- [**Disk and File Management**](../file/)
- [**Memory Management**](../buffer/)
- [**Transaction Management**](../transaction/)
- [**Record Management**](../record/)
- [**Metadata Management**](../metadata/)
- [**Query Processing**](../query/)
- [**Parsing Query**](../parse/)
- [**Planning Query**](../plan/)
- Continue...