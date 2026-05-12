/**
 * Маска телефона: 8 (927) 035-71-34
 * Поля с классом .phone-input получают форматирование при вводе.
 */
(function () {
  "use strict";

  var MAX_DIGITS = 11;

  function normalizeDigits(raw) {
    var d = String(raw).replace(/\D/g, "");
    if (d.length === 0) return "";
    if (d[0] === "7") d = "8" + d.slice(1);
    else if (d[0] !== "8") d = "8" + d;
    return d.slice(0, MAX_DIGITS);
  }

  function formatPhone(digits) {
    if (!digits || digits.length === 0) return "";
    var v = digits;
    var formatted = v[0];
    if (v.length > 1) {
      formatted += " (" + v.slice(1, Math.min(4, v.length));
    }
    if (v.length >= 4) {
      formatted += ")";
    }
    if (v.length > 4) {
      formatted += " " + v.slice(4, Math.min(7, v.length));
    }
    if (v.length > 7) {
      formatted += "-" + v.slice(7, Math.min(9, v.length));
    }
    if (v.length > 9) {
      formatted += "-" + v.slice(9, Math.min(11, v.length));
    }
    return formatted;
  }

  function onInput(e) {
    var input = e.target;
    var start = input.selectionStart;
    var prevLen = input.value.length;
    var normalized = normalizeDigits(input.value);
    var next = formatPhone(normalized);
    input.value = next;

    var diff = next.length - prevLen;
    if (typeof start === "number" && document.activeElement === input) {
      var pos = Math.max(0, Math.min(next.length, start + diff));
      try {
        input.setSelectionRange(pos, pos);
      } catch (err) {
        /* ignore */
      }
    }
  }

  function onPaste(e) {
    e.preventDefault();
    var text = (e.clipboardData || window.clipboardData).getData("text");
    var input = e.target;
    var normalized = normalizeDigits(text);
    input.value = formatPhone(normalized);
  }

  function initPhoneInputs(root) {
    var scope = root || document;
    var nodes = scope.querySelectorAll("input.phone-input");
    nodes.forEach(function (input) {
      if (input.dataset.mbPhoneInit === "1") return;
      input.dataset.mbPhoneInit = "1";
      input.setAttribute("inputmode", "numeric");
      input.setAttribute("autocomplete", "tel");
      input.addEventListener("input", onInput);
      input.addEventListener("paste", onPaste);
      if (input.value) {
        input.value = formatPhone(normalizeDigits(input.value));
      }
    });
  }

  if (document.readyState === "loading") {
    document.addEventListener("DOMContentLoaded", function () {
      initPhoneInputs(document);
    });
  } else {
    initPhoneInputs(document);
  }

  window.MorentBankPhoneMask = { initPhoneInputs: initPhoneInputs };
})();
