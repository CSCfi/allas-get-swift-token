# allas-get-swift-token
Allas SWIFT authentication utility application.

The application is a command line application especially for Windows users who want to use SWIFT environment variable authentication. When using this application in Windows, you (or your local it-support) may have to configure your anti-virus etc software to allow running allas-get-swift-token.

There are also easier ways to authenticate than this application, see https://docs.csc.fi/data/Allas/ for further details.

The program autehticates to CSC's Allas object storage and shows OS_AUTH_TOKEN and OS_STORAGE_URL values for swift applications to use, like for instance python-swiftclient, rclone or curl. For more information see https://docs.csc.fi/data/Allas/using_allas/rclone_local/#configuring-swift-connection-in-windows

## Compiling
You need the golang.org's terminal package to compile allas-get-swift-token: `golang.org/x/crypto/ssh/terminal`

The package is needed to ask password from the user without echoing the password while the user types it. That package needs a package that also might not be installed: `golang.org/x/sys/unix`

To install these two packages:
```
go get "golang.org/x/crypto/ssh/terminal"
go get "golang.org/x/sys/unix"
```
