// This file is based on fetch-event-stream package version 0.1.5.
// https://github.com/lukeed/fetch-event-stream
// This file was packaged using esbuild.

// MIT License
//
// Copyright (c) Luke Edwards <luke.edwards05@gmail.com> (lukeed.com)
//
// Permission is hereby granted, free of charge, to any person obtaining a copy of this software and associated documentation files (the "Software"), to deal in the Software without restriction, including without limitation the rights to use, copy, modify, merge, publish, distribute, sublicense, and/or sell copies of the Software, and to permit persons to whom the Software is furnished to do so, subject to the following conditions:
//
// The above copyright notice and this permission notice shall be included in all copies or substantial portions of the Software.
//
// THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY, FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM, OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE SOFTWARE.

// MIT License
//
// Copyright 2018-2022 the Deno authors.
//
// Permission is hereby granted, free of charge, to any person obtaining a copy
// of this software and associated documentation files (the "Software"), to deal
// in the Software without restriction, including without limitation the rights
// to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
// copies of the Software, and to permit persons to whom the Software is
// furnished to do so, subject to the following conditions:
//
// The above copyright notice and this permission notice shall be included in all
// copies or substantial portions of the Software.
//
// THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
// IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
// FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
// AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
// LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
// OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE
// SOFTWARE.

// node_modules/fetch-event-stream/esm/deps/jsr.io/@std/streams/0.221.0/text_line_stream.js
var TextLineStream = class extends TransformStream {
  #currentLine = "";
  /** Constructs a new instance. */
  constructor(options = { allowCR: false }) {
    super({
      transform: (chars, controller) => {
        chars = this.#currentLine + chars;
        while (true) {
          const lfIndex = chars.indexOf("\n");
          const crIndex = options.allowCR ? chars.indexOf("\r") : -1;
          if (crIndex !== -1 && crIndex !== chars.length - 1 && (lfIndex === -1 || lfIndex - 1 > crIndex)) {
            controller.enqueue(chars.slice(0, crIndex));
            chars = chars.slice(crIndex + 1);
            continue;
          }
          if (lfIndex === -1)
            break;
          const endIndex = chars[lfIndex - 1] === "\r" ? lfIndex - 1 : lfIndex;
          controller.enqueue(chars.slice(0, endIndex));
          chars = chars.slice(lfIndex + 1);
        }
        this.#currentLine = chars;
      },
      flush: (controller) => {
        if (this.#currentLine === "")
          return;
        const currentLine = options.allowCR && this.#currentLine.endsWith("\r") ? this.#currentLine.slice(0, -1) : this.#currentLine;
        controller.enqueue(currentLine);
      }
    });
  }
};

// node_modules/fetch-event-stream/esm/utils.js
function stream(input) {
  let decoder = new TextDecoderStream();
  let split2 = new TextLineStream({ allowCR: true });
  return input.pipeThrough(decoder).pipeThrough(split2);
}
function split(input) {
  let rgx = /[:]\s*/;
  let match = rgx.exec(input);
  let idx = match && match.index;
  if (idx) {
    return [
      input.substring(0, idx),
      input.substring(idx + match[0].length)
    ];
  }
}
function fallback(headers, key, value) {
  let tmp = headers.get(key);
  if (!tmp)
    headers.set(key, value);
}

// node_modules/fetch-event-stream/esm/mod.js
async function* events(res, signal) {
  if (!res.body)
    return;
  let iter = stream(res.body);
  let line, reader = iter.getReader();
  let event;
  for (; ; ) {
    if (signal && signal.aborted) {
      return reader.cancel();
    }
    line = await reader.read();
    if (line.done)
      return;
    if (!line.value) {
      if (event)
        yield event;
      event = void 0;
      continue;
    }
    let [field, value] = split(line.value) || [];
    if (!field)
      continue;
    if (field === "data") {
      event ||= {};
      event[field] = event[field] ? event[field] + "\n" + value : value;
    } else if (field === "event") {
      event ||= {};
      event[field] = value;
    } else if (field === "id") {
      event ||= {};
      event[field] = +value || value;
    } else if (field === "retry") {
      event ||= {};
      event[field] = +value || void 0;
    }
  }
}
async function stream2(input, init) {
  let req = new Request(input, init);
  fallback(req.headers, "Accept", "text/event-stream");
  fallback(req.headers, "Content-Type", "application/json");
  let r = await fetch(req);
  if (!r.ok)
    throw r;
  return events(r, req.signal);
}
export {
  events,
  stream2 as stream
};
