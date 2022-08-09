# Termsuji

termsuji is an application to play Go in your terminal. It is limited in features and scope, but you can play and finish games in progress on it. It is on github as a reference implementation, and not a reliable or stable package.

The *api* package can be used as a starting point to work with the online-go.com REST and realtime APIs. It only exports a limited part of the API and it may change without notice.

If you want to build yourself (or if your architecture isn't listed), [download/install Go](https://go.dev/dl), register an Oauth application at https://online-go.com/oauth2/applications/, place the client ID in `api/client_id.txt` without any whitespace, and run/build the application with `go run .` or `go build .`

![termsuji3](https://user-images.githubusercontent.com/110688516/183505721-6e50c05d-2572-4bb0-a06d-eae3006414a3.png)
![termsuji4](https://user-images.githubusercontent.com/110688516/183740301-19c66b74-d0ba-4fc2-a380-c9a3c08632e1.png)
