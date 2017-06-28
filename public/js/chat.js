/*
 * Editing pad component
 */

var ignoreCount = 0;

$(function() {
  var $chat = $("#chat");
  var sessionId = $chat.data("session");
  var documentId = $chat.data("document");
  var $messages = $chat.find('#messages');

  var nextMessage = 0;

  $chat.find('form').submit(function(e) {
    e.preventDefault();
    console.log('Form submitted');
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
        // TODO: Apply events
        for (i in events) {
          var ev = events[i];
          for (j in ev.messages) {
            var mes = ev.messages[j];
            $messages.append('<p><strong>' + mes.SessionID + '</strong>: ' + mes.Message + '</p>');
          }
        }
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
