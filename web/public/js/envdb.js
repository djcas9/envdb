var Envdb = {
  Socket: null,
  Init: function() {
    gotalk.handleNotification('results', function(data) {

      console.log(data)

      for (var i = 0, len = data.length; i < len; i++) {
        var agent = data[i];
        agent.results = JSON.parse(agent.results)
        console.log(agent)
      }
    });

    Envdb.Socket = gotalk.connection().on('open', function() {
      // ..
    });
  }
};

jQuery(document).ready(function($) {

  // BroTop.SetVersion();

  Envdb.Init()

  $('#search').keyup(function(e) {
    var self = this;

    if (e.keyCode == 13) {
      var value = $(self).val();
      console.log(value)

      $("div.results").html("")
      $("div.results").append("Loading...")

      Envdb.Socket.request('query', {
        id: "all",
        sql: value,
      }, function (err, data) {

        console.log("BEFORE ERROR::", data)

        $("div.results").html("")

        for (var i = 0, len = data.length; i < len; i++) {
          var agent = data[i];
          agent.results = JSON.parse(agent.results)
          console.log(agent.results)

          $("div.results").append("<br /><br />")
          $("div.results").append("<strong>Agent: "+agent.name+" :: "+agent.id+"</strong>")
          $("div.results").append("<br /><br />")
          $("div.results").append(JSON.stringify(agent))
          $("div.results").append("<br /><br />")

        }
      });
    }
  });

  $('#search2').keyup(function(e) {
    var self = this;

    if (e.keyCode == 13) {
      var value = $(self).val();
      console.log(value)

      $("div.results").html("")
      $("div.results").append("Loading...")

      Envdb.Socket.request('tables', {
        id: value
      }, function (err, data) {

        console.log("word?")
        console.log(err, data)

        $("div.results").html(data.results)

      });
    }
  });

  $('#search3').keyup(function(e) {
    var self = this;

    if (e.keyCode == 13) {
      var value = $(self).val();
      var id = $("#search2").val();
      console.log(value)

      $("div.results").html("")
      $("div.results").append("Loading...")

      Envdb.Socket.request('table-info', {
        id: id,
        sql: "pragma table_info("+value+");",
      }, function (err, data) {

        console.log("word?")
        console.log(err, data)

        $("div.results").html(JSON.stringify(data))

      });
    }
  });

});
