/*

  https://github.com/matteobad/detect-autofill

  MIT License

  Copyright (c) 2019 Matteo Badini

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

import "./custom-event-polyfill";
import "./detect-autofill.scss";

document.addEventListener("animationend", onAnimationStart, true);
document.addEventListener("input", onInput, true);

/**
 * Handler for -webkit based browser that listen for a custom
 * animation create using the :pseudo-selector in the stylesheet.
 * Works with Chrome, Safari
 *
 * @param {AnimationEvent} event
 */
function onAnimationStart(event) {
  "onautofillend" === event.animationName
    ? autocomplete(event.target)
    : cancelAutocomplete(event.target);
}

/**
 * Handler for non-webkit based browser that listen for input
 * event to trigger the autocomplete-cancel process.
 * Works with Firefox, Edge, IE11
 *
 * @param {InputEvent} event
 */
function onInput(event) {
  "insertReplacementText" === event.inputType || !("data" in event)
    ? autocomplete(event.target)
    : cancelAutocomplete(event.target);
}

/**
 * Manage an input element when its value is autocompleted
 * by the browser in the following steps:
 * - add [autocompleted] attribute from event.target
 * - create 'onautocomplete' cancelable CustomEvent
 * - dispatch the Event
 *
 * @param {HtmlInputElement} element
 */
function autocomplete(element) {
  if (element.hasAttribute("autocompleted")) return;
  element.setAttribute("autocompleted", "");

  var event = new window.CustomEvent("onautocomplete", {
    bubbles: true,
    cancelable: true,
    detail: null,
  });

  // no autofill if preventDefault is called
  if (!element.dispatchEvent(event)) {
    element.value = "";
  }
}

/**
 * Manage an input element when its autocompleted value is
 * removed by the browser in the following steps:
 * - remove [autocompleted] attribute from event.target
 * - create 'onautocomplete' non-cancelable CustomEvent
 * - dispatch the Event
 *
 * @param {HtmlInputElement} element
 */
function cancelAutocomplete(element) {
  if (!element.hasAttribute("autocompleted")) return;
  element.removeAttribute("autocompleted");

  // dispatch event
  element.dispatchEvent(
    new window.CustomEvent("onautocomplete", {
      bubbles: true,
      cancelable: false,
      detail: null,
    })
  );
}
