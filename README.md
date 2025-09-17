# jira-implementation-generator

jira-implementation-generator (jig) is a tool that helps generate an implementation plan for a Jira ticket,
utilizing the Google Vertix AI API. By fetching the ticket, feeding that data into a prompt, and passing it to a model
like `claude-sonnet-4`, you can generate a comprehensive plan that includes things like:

* What is the ticket asking for?
* What repos might be affected by the change?
* What considerations do you need to take when making changes?
* A _rough_ estimate on difficulty/time it will take to implement.

## Running

Build the tool with `make build`, then you can run it:

    jig --help

By default, the configuration will use issues.redhat.com, and your local Google Cloud credentials to connect to a Google
Vertex AI API and project. To customize this, pass the relevant flags:

    jig --token $JIRA_TOKEN --region $REGION --project-id $PROJECT_ID --jira-base-url jira.com $TICKET_ID
