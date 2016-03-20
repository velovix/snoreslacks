# Snoreslacks
Snoreslacks is a slash command server for Slack that allows users to engage in
Pokemon battles with each other.

## Set Up
Snoreslacks is built for Google App Engine. You can create a free instance
at the time of this writing on Google Cloud Platform. As long as your server
doesn't get too much traffic, that should be sufficient.

Before deploying your app, you need to add a `snoreslacks.yaml` to the `/app`
directory. It only needs to contain a token field so that Snoreslacks can verify
that requests are coming from your channel. Here is an example.
