/*
 * Editing pad component
 */

var ignoreCount = 0;

$(function() {
  var editor = ace.edit("editor");
  var sessionId = $("#editor").data("session");
  var nextMessage = 0;

  editor.$blockScrolling = Infinity
  editor.setTheme("ace/theme/monokai");
  editor.getSession().setMode("ace/mode/javascript");
  var document = editor.getSession().getDocument();
  editor.on('change', function(e) {
    if (ignoreCount > 0) {
      ignoreCount--;
      return;
    }

    $.ajax({
      url: "/push",
      type: "POST",
      data: JSON.stringify({
        sessionId: "" + sessionId,
        deltas: [e],
      }),
      contentType: "application/json",
      complete: function(result) {
        // TODO: Handle the response
      }
    });
  });

  editor.getSession().selection.on('changeSelection', function(e) {
    // TODO: Use editor.getSelectionRange() to update selection shown.
  });

  editor.getSession().selection.on('changeCursor', function(e) {
    // TODO: Use editor.selection.getCursor() to update cursor position
  });

  var pollUpdates = function() {
    $.ajax({
      url: "/poll",
      type: "GET",
      data: {
        sessionId: "" + sessionId,
        next: nextMessage,
      },
      success: function(result) {
        events = JSON.parse(result);
        if (!events) {
          pollUpdates();
          return;
        }
        deltas = []
        for (i in events) {
          var e = events[i];
          if(e.sessionId == sessionId) {
            continue;
          }
          deltas = deltas.concat(e.deltas)
        }
        ignoreCount = deltas.length;
        document.applyDeltas(deltas)
        nextMessage+=events.length;
        pollUpdates();
      },
      error: function(result, text, errorThrown) {
        if (result.readyState != 0) {
          console.log("Poll cancelled");
          return;
        }
        console.log("Error: " + text);
        console.log(result);
        console.log(errorThrown);
        pollUpdates();
      }
    });
  };
  pollUpdates();
});
