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
    console.log(editor.getSelectionRange());
  });

  editor.getSession().selection.on('changeCursor', function(e) {
    console.log(editor.selection.getCursor());
  });

  var pollUpdates = function() {
    $.ajax({
      url: "/poll",
      type: "GET",
      data: {
        sessionId: "" + sessionId,
        next: nextMessage,
      },
      complete: function(result) {
        events = JSON.parse(result.responseText);
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
      }
    });
  };
  pollUpdates();
});
