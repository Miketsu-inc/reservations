/*

  Polyfill for creating CustomEvents on IE9/10/11

  code pulled from:
  https://github.com/d4tocchini/customevent-polyfill
  https://developer.mozilla.org/en-US/docs/Web/API/CustomEvent#Polyfill

  The MIT License (MIT)

  Copyright (c) 2016 Evan Krambuhl

  Permission is hereby granted, free of charge, to any person obtaining a copy
  of this software and associated documentation files (the "Software"), to deal
  in the Software without restriction, including without limitation the rights
  to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
  copies of the Software, and to permit persons to whom the Software is
  furnished to do so, subject to the following conditions:

  The above copyright notice and this permission notice shall be included in all
  copies or substantial portions of the Software.

  THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
  IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
  FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
  AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
  LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
  OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE
  SOFTWARE.

*/

(function () {
  if (typeof window === "undefined") {
    return;
  }

  try {
    var ce = new window.CustomEvent("test", { cancelable: true });
    ce.preventDefault();
    if (ce.defaultPrevented !== true) {
      // IE has problems with .preventDefault() on custom events
      // http://stackoverflow.com/questions/23349191
      throw new Error("Could not prevent default");
    }
  } catch (_) {
    var CustomEvent = function (event, params) {
      var evt, origPrevent;

      // We use here some version of `Object.assign` implementation, to create a shallow copy of `params`.
      // Based on https://github.com/christiansany/object-assign-polyfill/blob/213cc63df14515fb543117059d1576204bfaa8a7/index.js
      var newParams = {};
      // Skip over if undefined or null
      if (params != null) {
        for (var nextKey in params) {
          // Avoid bugs when hasOwnProperty is shadowed
          if (Object.prototype.hasOwnProperty.call(params, nextKey)) {
            newParams[nextKey] = params[nextKey];
          }
        }
      }

      newParams.bubbles = !!newParams.bubbles;
      newParams.cancelable = !!newParams.cancelable;

      evt = document.createEvent("CustomEvent");
      evt.initCustomEvent(
        event,
        newParams.bubbles,
        newParams.cancelable,
        newParams.detail
      );
      origPrevent = evt.preventDefault;
      evt.preventDefault = function () {
        origPrevent.call(this);
        try {
          Object.defineProperty(this, "defaultPrevented", {
            get: function () {
              return true;
            },
          });
        } catch (_) {
          this.defaultPrevented = true;
        }
      };
      return evt;
    };

    CustomEvent.prototype = window.Event.prototype;
    window.CustomEvent = CustomEvent; // expose definition to window
  }
})();
