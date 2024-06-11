# Chat App

Using Pocketbase, Templ, HTMX, Alpine.js, and Tailwind CSS

## Installing and running
- Run `git clone https://github.com/charliekim2/chatapp.git`
- `cd` into the directory
- To run a server with hot reload:
	- Install templ via `go install github.com/a-h/templ/cmd/templ@latest`
	- Install air via `go install github.com/air-verse/air@latest`
	- Run `air` to start the server
	- Go to the admin dashboard via `localhost:8090/_` and import `pb_schema.json`
- Or, if not running via air, make sure to `templ generate` the template files
