# Core

## Features

1. Set up a time, for example, 30 seconds. The client initiates this. The server will then generate a UUID to be associated with the "session".
2. During that 30 seconds period, grab a word from the **bucket**.
   1. This is retrieved from SSE, not client-side. The SSE is sent according to the UUID.
   2. The SSE also sends a unique "token" which is responsible to determine whether the request comes from the server or from elsewhere (e.g. script).
3. Scramble the order of letters of the bucket. Scramble the word 10 times and then get the lowest edit distance. Send the word via SSE.
4. When the player submits a correct answer, the UI Sends a HTTP request. The server will count this as correct answer.
   1. If the answer is correct AND contains the said "token", then the request is valid.
   2. If the answer is incorrect OR does not contain the said "token", then the request is invalid.
5. When the time ends, count the number of words that were successful to be re-ordered.
   1. The timer "end status" will come from SSE. The server will send the number of correct answers.
6. Store this in a SQLite database.

## Behind the scenes features

1. Find a way to scrape Merriam-Webster for 5-letter words, A-Z. These will go into a something called a **Bucket**.

# MVP Features

1. Implement the "Behind the scenes" features.
2. Implement all the features above client-side. There is no need for SSE or HTTP. For SQLite, store it in-memory.
