/*
 * Editing pad component
 */

var ignoreEdit = false;

$(function() {
  var editor = ace.edit("editor");
  var sessionId = $("#editor").data("session");

  editor.$blockScrolling = Infinity
  editor.setTheme("ace/theme/monokai");
  editor.getSession().setMode("ace/mode/javascript");
  var document = editor.getSession().getDocument();
  editor.on('change', function(e) {
    if (ignoreEdit) {
      ignoreEdit = false;
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
      },
      complete: function(result) {
        console.log("Poll response");
        event = JSON.parse(result.responseText);
        ignoreEdit = true;
        document.applyDeltas(event.deltas);
        pollUpdates();
      }
    });
  };
  pollUpdates();
});
