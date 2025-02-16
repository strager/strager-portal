import * as fetchEventStream from "./thirdparty/fetch-event-stream.mjs";

// TODO(strager): Move this out of the repo.
let groqAPIKey = "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";

function redirectToKagi(query) {
    console.log("redirectToKagi(" + JSON.stringify(query) + ")");
    showRedirectingUI("Kagi");
    window.location = "https://kagi.com/search?" + new URLSearchParams({"q": query});
}

function showRedirectingUI(destination) {
    // @type DOMElement | null
    let redirectingElement = document.getElementById("redirecting");

    // Delay showing the UI to avoid a flash of content.
    setTimeout(() => {
        redirectingElement.style.display = "";
    }, 250);
}

async function initiateAIConversationAsync(query) {
    // @type DOMElement | null
    let chatElement = document.getElementById("chat");
    chatElement.style.display = "";

    let userMessageElement = document.createElement("div");
    userMessageElement.classList.add("user");
    userMessageElement.textContent = query;
    chatElement.appendChild(userMessageElement);

    let assistantMessageElement = document.createElement("div");
    assistantMessageElement.classList.add("assistant");
    chatElement.appendChild(assistantMessageElement);

    for await (let data of await aiAsync([query])) {
        assistantMessageElement.appendChild(document.createTextNode(data));
    }
}

// @param conversation Array<string>
// @return AsyncIterator<ServerSentEventMessage>
async function* aiAsync(conversation, abortSignal = null) {
    // @type Array
    let messages = [
        {
            role: "system",
            content: "You are an assistant helping an experienced software engineer. The engineer is requesting information. Please provide the information requested in a concise manner without headings or unnecessary explanation. If appropriate, show a short code example in the language mentioned. Keep commentary to a minimum. Unless requested, do not include error handling code.",
        },
    ];
    for (let i = 0; i < conversation.length; i += 1) {
        messages.push({
            role: i % 2 === 0 ? "user" : "assistant",
            content: conversation[i],
        });
    }
    let response = await fetch("https://api.groq.com/openai/v1/chat/completions", {
        method: "POST",
        headers: {
            "Content-Type": "application/json",
            "Authorization": `Bearer ${groqAPIKey}`,
        },
        body: JSON.stringify({
            messages: messages,
            model: "llama-3.3-70b-versatile",
            max_completion_tokens: 1000,
            stream: true,
        }),
        signal: abortSignal,
    });
    if (!response.ok) {
        throw new Error("Request failed");
    }

    let stream = fetchEventStream.events(response, abortSignal);
    for await (let event of stream) {
        if (event.data === "[DONE]") {
            break;
        }
        let completion = JSON.parse(event.data);
        let content = completion.choices[0].delta?.content ?? "";
        if (content !== "") {
            yield content;
        }
    }
}

function main() {
    // @type string
    let query = new URLSearchParams(document.location.search).get("q") ?? "";
    // @type Array<string>
    let tokens = query.split(/\s+/g);
    // @type string | null
    let bangToken = null;
    if (tokens.length > 0) {
        if (tokens[0].startsWith("!")) {
            bangToken = tokens[0];
        } else if (tokens.at(-1).startsWith("!")) {
            bangToken = tokens.at(-1);
        }
    }

    if (bangToken !== null) {
        redirectToKagi(query);
        return;
    }
    if (query.endsWith("?")) {
        initiateAIConversationAsync(query);
        return;
    }
    redirectToKagi(query);
}

main()
