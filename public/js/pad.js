/*
 * Editing pad component
 */

$(function() {
  $('#pad').get().hideFocus = true;
  $('#pad').focus();
});

$("#pad").on('keydown', function(e) {
  var cursor = $(this).find("#cursor");
  var keyCode = e.keyCode || e.which;
  // Modifier keys
  if (e.ctrlKey || e.metaKey || e.shiftKey || e.altKey) {
    return;
  }
  // Function Keys
  if (keyCode >= 112 && keyCode <= 123) {
    return;
  }
  switch (keyCode) {
    case 16:
    case 27:
      break;
    case 8: // Delete
      e.preventDefault();
      doDelete(cursor);
      break;
    case 32: // Space
      cursor.before('<div class="char">&nbsp;</div>');
      break;
    case 37: // Left
      doLeft(cursor);
      break;
    case 39: // Right
      doRight(cursor);
      break;
    case 38: // Up
      doUp(cursor);
      break;
    case 40: // Down
      doDown(cursor);
      break;
    case 9: // Tab
      e.preventDefault();
      cursor.before('<div class="char">&emsp;</div>');
      break;
    case 13:
      e.preventDefault();
      doNewLine(cursor);
      break;
    default:
      cursor.before('<div class="char">' + event.key + "</div>");
      console.log(event);
      break;
  }
});

$('#pad').on('click', function(){
  var mytext = selectHTML();
  $('span').css({"background":"yellow","font-weight":"bold"});
});

function doNewLine(cursor) {
  line = currentLine(cursor);
  line.after('<div class="line"></div>');
  line.next('.line').append(cursor.nextAll().detach());
  cursor = moveToStart(cursor, line.next());
}

function doLeft(cursor) {
  if (cursor.prev().length !== 0) {
    var div2 = cursor.prev().detach();
    cursor.after(div2);
    return;
  }
  line = currentLine(cursor);
  previousLine = line.prev('.line');
  if (previousLine.length === 0) {
    return;
  }
  moveToEnd(cursor, previousLine);
}

function doRight(cursor) {
  if (cursor.next().length !== 0) {
    var div2 = cursor.next().detach();
    cursor.before(div2);
    return;
  }
  line = currentLine(cursor);
  nextLine = line.next('.line');
  if (nextLine.length === 0) {
    return;
  }
  moveToStart(cursor, nextLine);
}

function doDelete(cursor) {
  if (cursor.prev().length !== 0) {
    cursor.prev().remove();
    return;
  }

  line = currentLine(cursor);
  previousLine = line.prev('.line');
  if (previousLine.length === 0) {
    return;
  }
  moveToEnd(cursor, previousLine);
  if (line.children().length === 0) {
    line.remove();
  }
}

function doUp(cursor) {
  line = currentLine(cursor);
  pos = positionInLine(cursor);
  previousLine = line.prev('.line');
  if (previousLine.length === 0) {
    previousLine = line;
  }
  moveToStart(cursor, previousLine);
  moveToPositionInLine(cursor, pos);
}

function doDown(cursor) {
  line = currentLine(cursor);
  pos = positionInLine(cursor);
  nextLine = line.next('.line');
  if (nextLine.length === 0) {
    nextLine = line;
  }
  moveToStart(cursor, nextLine);
  moveToPositionInLine(cursor, pos);
}

function currentLine(cursor) {
  return cursor.closest('.line');
}

function moveToPositionInLine(cursor, pos) {
  console.log("Moving to position: " + pos);
  var line = currentLine(cursor);
  cursor = moveToStart(cursor, line);
  for(var i = 0; i < pos; i++){
    var div2 = cursor.next().detach();
    cursor.before(div2);
  }
}

function positionInLine(cursor) {
  var line = currentLine(cursor);
  var children = line.children();
  for(var i = 0; i < children.length; i++){
    if ($(children[i]).hasClass('js-cursor')) {
      return i;
    }
  }
  return 0;
}

function moveToStart(cursor, target) {
  if(target == cursor) {
    return;
  }
  c = cursor.detach();
  target.prepend(c);
  return cursor
}

function moveToEnd(cursor, target) {
  if(target == cursor) {
    return;
  }
  c = cursor.detach();
  target.append(c);
  return cursor
}

function moveBefore(cursor, targetChar) {
  if(targetChar == cursor) {
    return;
  }
  c = cursor.detach();
  targetChar.before(c);
  return c;
}

function moveAfter(cursor, targetChar) {
  if(targetChar == cursor) {
    return;
  }
  c = cursor.detach();
  targetChar.after(c);
  return c;
}

// Example of wrapping selected text
function selectHTML() {
    try {
        if (window.ActiveXObject) {
            var c = document.selection.createRange();
            return c.htmlText;
        }

        var nNd = document.createElement("span");
        var w = getSelection().getRangeAt(0);
        w.surroundContents(nNd);
        return nNd.innerHTML;
    } catch (e) {
        if (window.ActiveXObject) {
            return document.selection.createRange();
        } else {
            return getSelection();
        }
    }
}
