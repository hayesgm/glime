(function() {
  var settings = {
    style: "position:absolute;top:0;left:0",
    width: $(window).width(),
    height: $(window).height()
  };

  var dbg = function(msg) {
    if (typeof(console) !== 'undefined' && console.log && location.hash == "#debug") {
      console.log(msg);
    }
  };

  var GameState = {
    objects: {},
    messages: []
  };

  var printMessage = function(message) {
    GameState.messages.unshift(message);
    GameState.messages = GameState.messages.slice(0, 10);
  }

  // Game state is going to be a collection of GameObjects
  // We are going to, at first, just get updates on game-state
  // We are only going to use deltas, and we'll use hash functions to generate ids
  var GameMessage = function(type, payload) {
    this.type = type;
    this.payload = JSON.stringify(payload);
  }

  GameMessage.prototype.send = function(server) {
    var message = JSON.stringify(this);

    setTimeout(function() {
      //dbg(['Sending message', message]);
      server.send(message);
    },0); // Async
  }

  var GameObject = function(id, ownerId, x, y, bearing, objType, qualities, characteristics, remove) {
    this.id = id;
    this.ownerId = ownerId;
    this.x = x;
    this.y = y;
    this.bearing = bearing;
    this.objType = objType;
    this.qualities = qualities;
    this.characteristics = characteristics;
    this.remove = remove
  };

  GameObject.prototype.toString = function() {
    return "GameObject: {id:" + this.id + ", ownerId: " + this.ownerId + ", x:" + this.x + ", y:" + this.y + ", type" + this.type + ", qualities: " + this.qualities + ", characteristics: " + this.characteristics + "}";
  }

  var buildGameObject = function(json) {
    dbg(['Building game object from', json]);
    var obj = new GameObject(json.Id, json.OwnerId, json.X, json.Y, json.Bearing, json.ObjType, json.Qualities, json.Characteristics, json.Remove);
    dbg(['Built',obj]);
    return obj;
  };

  GameObject.prototype.clear = function(ctx) {
    switch (this.objType) {
    case "Player":
      ctx.strokeStyle = "#000";
      ctx.fillStyle = "#000";

      ctx.beginPath();
      ctx.arc(parseInt(this.x), parseInt(this.y), parseInt(this.characteristics.width + 2.0), 0, Math.PI*2, true); 
      ctx.closePath();
      ctx.fill();
      break;
    case "Projectile":
      ctx.strokeStyle = "#000";
      ctx.fillStyle = "#000";

      ctx.beginPath();
      ctx.arc(parseInt(this.x), parseInt(this.y), parseInt(this.characteristics.width + 2.0), 0, Math.PI*2, true); 
      ctx.closePath();
      ctx.fill();
      break;
    default:
      dbg(["Don't know how to clear " + this.objType, this])
    }
  };

  GameObject.prototype.draw = function(ctx) {
    ctx.save();

    switch (this.objType) {
    case "Player":
      ctx.strokeStyle = "#000";
      ctx.fillStyle = this.qualities.color;
      ctx.beginPath();
      ctx.arc(parseInt(this.x), parseInt(this.y), this.characteristics.width, 0, Math.PI*2, true); 
      ctx.closePath();
      ctx.fill();

      if (this.bearing) {
        // dbg(['Drawing bearing', this.x, this.y, this.x + 10 * Math.sin(this.bearing), this.y + 10 * Math.cos(this.bearing)]);
        ctx.strokeStyle = "#eee";
        ctx.fillStyle = "#000";
        ctx.moveTo(this.x + 0.3 * this.characteristics.width * Math.cos(this.bearing), this.y + 0.3 * this.characteristics.width * Math.sin(this.bearing));
        ctx.lineTo(this.x + this.characteristics.width * Math.cos(this.bearing), this.y + this.characteristics.width * Math.sin(this.bearing));
        ctx.stroke();
      }
      break;
    case "Projectile":
      ctx.strokeStyle = "#000";
      ctx.fillStyle = this.qualities.color;
      ctx.beginPath();
      ctx.arc(parseInt(this.x), parseInt(this.y), this.characteristics.width, 0, Math.PI*2, true);
      ctx.closePath();
      ctx.fill();
      break;
    default:
      dbg(["Don't know how to draw " + this.objType, this])
    }

    ctx.restore();
  };

  GameObject.prototype.calculateBearing = function(targetX, targetY) {
    var dx = targetX - this.x;
    var dy = targetY - this.y; // This is fipped in our coordinate system
    return this.bearing = Math.atan2(dy, dx);
  }

  GameObject.prototype.move = function(dx, dy) {
    this.x += dx;
    this.y += dy;
  };
  
  $(document).ready(function() {
    $('canvas[data-game=glime]').each(function(){
      // Okay, we're in a glime session now
      var $self = $(this);
      var canvas = $self.get(0);
      var ctx = canvas.getContext("2d");
      var objects = [];
      var targetX = null;
      var targetY = null;
      var playerId = null;

      $self.attr(settings); // Adjust height and width

      ctx.fillStyle = "#000";
      ctx.fillRect(0, 0, settings.width, settings.height);

      var addObject = function(object) {
        objects.push(object);
        return object;
      };

      var clearScreen = function() {
        /*
        ctx.fillStyle = "#000";
        for (var i in GameState.objects) {
          // dbg(['Drawing', i, GameState.objects[i].toString()]);
          var obj = GameState.objects[i];
          obj.clear(ctx);
        }
        */
        ctx.fillStyle = "#000";
        ctx.fillRect(0, 0, settings.width, settings.height);
      };

      var drawObjects = function() {
        for (var i in GameState.objects) {
          // dbg(['Drawing', i, GameState.objects[i].toString()]);
          var obj = GameState.objects[i];
          obj.draw(ctx);
        }
      };

      var drawMessages = function() {
        var x = 300;
        var y = 0;
        for (var k in GameState.messages) {
          var message = GameState.messages[k];
          ctx.font = "bold 12px sans-serif";
          ctx.stoke = "#ccc";
          ctx.fillText(message, x, y);
          y += 20;
        }
      };

      var server = new WebSocket($self.data('src'),['xmpp']);
      dbg(['Connecting to', $self.data('src')]);

      server.onmessage = function(e) {
        var json = JSON.parse(e.data);

        var obj = buildGameObject(json);

        if (playerId == null) { // First object is always us
          playerId = obj.id;
          dbg(['I am', playerId]);
        }

        dbg(['Update world', obj]);
          
        // These are all object updates, so we're simply going to fill it into our known object universe
        // Adding it if need be
        if (GameState.objects[obj.id]) {
          GameState.objects[obj.id].clear(ctx);
        }
        if (obj.remove) {
          delete GameState.objects[obj.id];
        } else {
          GameState.objects[obj.id] = obj; // TODO: We may want to cover out-of-order messages
          obj.draw(ctx);
        }

        dbg(['The world',GameState]);
      }

      server.onopen = function() { // We'll wait till the connection is opened to continue processing
        dbg(['Connected to server']);

        var gameLoop = function() { // This will be running continuously at some interval
          //clearScreen();
          drawObjects();
          drawMessages();
        };

        // var player = addObject(new GameObject(20, 20)); // Load the initial game
        var registerListeners = function() {
          document.onmousemove = function(e) {
            targetX = e.x;
            targetY = e.y;
            // var player = GameState.objects[playerId];
            // printMessage('(tX: ' + targetX + ', tY: ' + targetY + "), (" + player.x + ', ' + player.y + '), ' + ( player.bearing(targetX, targetY) / Math.PI ) + 'PI');
          };
          document.onmousedown = function() {
            if (playerId == null) { // We can't accept options until initialization of player
              return;
            }

            var bearing = GameState.objects[playerId].calculateBearing(targetX, targetY);
            new GameMessage('fire', {bearing: bearing}).send(server)
            // var player = GameState.objects[playerId];
            // dbg(['px', player.x, 'targetX', targetX, 'py', player.y, 'targetY', targetY]);
          };

          document.onkeydown = function(e) {
            if (playerId == null) { // We can't accept options until initialization of player
              return;
            }

            e = e || window.event;
            dbg(['Find bearing from',playerId,GameState.objects[playerId]]);
            var bearing = GameState.objects[playerId].calculateBearing(targetX, targetY);

            switch (parseInt(e.keyCode)) {
            case 38, 87: // up arrow
              dbg(['Move toward', playerId, bearing]);
              new GameMessage('move', {direction: 'toward', bearing: bearing}).send(server)
              break;
            case 40, 83: // down arrow
              dbg(['Move away', playerId, bearing]);
              new GameMessage('move', {direction: 'away', bearing: bearing}).send(server)
              break;
            case 37, 65: // left arrow
              dbg(['Strafe left', playerId, bearing]);
              new GameMessage('move', {direction: 'strafe-left', bearing: bearing}).send(server)
              break;
            case 39, 68: // right arrow
              dbg(['Strafe right', playerId, bearing]);
              new GameMessage('move', {direction: 'strafe-right', bearing: bearing}).send(server)
              break;
            default:
              dbg(['Ignoring key', e.keyCode]);
              return true;
            }

            e.preventDefault();
          };
        };

        setInterval(gameLoop, 200); // Game the game loop every so often
        registerListeners(); // Give us controls  
      }
      
      // Connecting...
    });
  });

})();