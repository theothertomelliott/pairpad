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
  $messages.text("");

  var $nameForm = $chat.find('#change_name');
  var $nameDisplay = $chat.find('#name_display');
  $nameForm.hide();
  $nameDisplay.find('#change_name_link').click(function(e) {
    e.preventDefault();
    $nameDisplay.hide();
    $nameForm.show();
    $nameForm.find('#new_name').val(sessionNames[sessionId]);
  });

  var $currentName = $chat.find('#current_name');
  var $users = $chat.find('#users');

  var nextMessage = 0;

  var addMessages = function(messages) {
    var atBottom = $messages.text().length == 0 || $messages.prop('scrollHeight') == $messages.scrollTop() + $messages.prop('clientHeight');
    for (j in messages) {
      var mes = messages[j];
      var name = sessionNames[mes.SessionID];
      if (!name) {
        name = mes.SessionID;
      }
      $messages.append('<p><strong>' + name + '</strong>: ' + mes.Message + '</p>');
    }
    if (atBottom) {
      $messages.animate({ scrollTop: $messages.prop("scrollHeight")}, 200);
    }
  };

  var updateUsers = function() {
    $currentName.text(sessionNames[sessionId]);
    $users.text("");
    for (id in sessionNames) {
      $users.append("<p>" + sessionNames[id] + "</p>");
    }
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
  $chat.find('#change_name').on('reset',function(e) {
    e.preventDefault();
    $nameDisplay.show();
    $nameForm.hide();
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
        var messages = [];
        for (i in events) {
          var ev = events[i];
          messages = messages.concat(ev.messages);
          for (key in ev.sessionNameChanges) {
            var value = ev.sessionNameChanges[key];
            sessionNames[key] = value;
          }
        }
        if (messages.length > 0) {
          addMessages(messages);
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
