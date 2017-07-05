/*
 * Editing pad component
 */

var ignoreCount = 0;

$(function() {
  var $chat = $("#chat");
  var sessionId = $chat.data("session");

  var sessionNames = {};
  sessionNames[sessionId] = sessionId;

  var documentId = $chat.data("document");
  var $messages = $chat.find('#messages');

  var $nameForm = $chat.find('#change_name');
  var $nameDisplay = $chat.find('#name_display');
  $nameForm.hide();
  $nameDisplay.click(function(e) {
    $nameDisplay.hide();
    $nameForm.show();
    $nameForm.find('#new_name').val(sessionNames[sessionId]);
  });

  var $currentName = $chat.find('#current_name');
  var $users = $chat.find('#users');

  var nextMessage = 0;

  var updateUsers = function() {
    $currentName.text(sessionNames[sessionId]);
    $users.text("");
    $users.append("<ul>");
    for (id in sessionNames) {
      $users.append("<li>" + sessionNames[id] + "</li>");
    }
    $users.append("</ul>");
  };
  updateUsers();

  $chat.find('#post_message').submit(function(e) {
    e.preventDefault();
    var $newMessage = $(this).find('#new_message');
    var msg = $newMessage.val();
    $.ajax({
      url: "/chat/push/" + documentId,
      type: "POST",
      data: JSON.stringify({
        messages: [{
          sessionId: "" + sessionId,
          message: msg
        }],
      }),
      contentType: "application/json",
      complete: function(result) {
        $newMessage.val("");
      }
    });
  });

  $chat.find('#change_name').submit(function(e) {
    e.preventDefault();
    var $newName = $(this).find('#new_name');
    var name = $newName.val();
    var sessionNameChanges = {}
    sessionNameChanges[sessionId] = name;
    $.ajax({
      url: "/chat/push/" + documentId,
      type: "POST",
      data: JSON.stringify({
        sessionNameChanges: sessionNameChanges
      }),
      contentType: "application/json",
      complete: function(result) {
        $nameDisplay.show();
        $nameForm.hide();
      }
    });
  });

  var pollUpdates = function() {
    $.ajax({
      url: "/chat/poll/" + documentId,
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
        for (i in events) {
          var ev = events[i];
          for (j in ev.messages) {
            var mes = ev.messages[j];
            var name = sessionNames[mes.SessionID];
            if (!name) {
              name = mes.SessionID;
            }
            $messages.append('<p><strong>' + name + '</strong>: ' + mes.Message + '</p>');
          }
          for (key in ev.sessionNameChanges) {
            var value = ev.sessionNameChanges[key];
            sessionNames[key] = value;
          }
        }
        updateUsers();
        nextMessage+=events.length;
        pollUpdates();
      },
      error: function(result, text, errorThrown) {
        if (result.readyState == 0) {
          // TODO: Display "connection lost" or similar
          return;
        }
        // Retry after 1s
        setTimeout(pollUpdates, 1000);
      }
    });
  };
  pollUpdates();
});
