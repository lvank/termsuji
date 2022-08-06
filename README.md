# Termsuji

termsuji is an application to play Go in your terminal. It is limited in features and scope and cannot yet play full games or issue challenges. Passing turns, for instance, is not possible. It is on github as a reference implementation, and not a reliable or stable package.

The *api* package can be used as a starting point to work with the online-go.com REST and realtime APIs. It only exports a limited part of the API and it may change without notice.

To run, register an Oauth application at https://online-go.com/oauth2/applications/, place the client ID in `api/client_id.txt` without any whitespace, and run/build the application.