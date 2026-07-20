

class Promise {
  private default promise;
  private thenHandlers = [];
  private catchHandlers = [];
  private state = "pending"; $ ("pending" | "unfulfilled" | "fulfilled")
  private result;

  constructor(callback) {
    if (typeof callback != "function") {
      throw #_new_error("TypeError", "Promise: 1st argument must be a function", {callsite: true});
    }
    try {
      callback()
    } catch (reason) {
      this.reject(reason)
    }
    //this.promise = #_new_promise(callback);
  }

  private resolve(value) {
    this.state = "fulfilled";
    this.result = value;
    for (let handler of thenHandlers) {
      handler(value)
    }
  }

  private reject(reason) {
    this.state = "unfulfilled";
    this.result = reason;
    for (let handler of catchHandlers) {
      handler(reason)
    }
  }

  then(callback) {
    if (typeof callback != "function") {
      throw #_new_error("TypeError", "Promise.then: 1st argument must be a function", {callsite: true});
    }
    // this.promise.then(callback)
    if (this.state == "fulfilled") callback(this.result);
    else {
      #_append(this.thenHandlers, callback)
    }
    return this
  }

  _catch(callback) {
    if (typeof callback != "function") {
      throw #_new_error("TypeError", "Promise._catch: 1st argument must be a function", {callsite: true});
    }
    // this.promise.catch(callback)
    if (this.state == "unfulfilled") callback(this.result);
    else {
      #_append(this.catchHandlers, callback)
    }
    return this
  }

  [Symbol.debug]() {
    let value = "\x1b[34m"+this.state+"\x1b[0m";
    if (this.state == "fulfilled") value = #_inspect(this.result);
    return "\x1b[32mPromise\x1b[0m { " + value + " }"
  }
}

