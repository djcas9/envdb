var Envdb = {
  table: false,
  fixedTable: false,

  Loading: {
    options: {
      ajax: true,
      document: true,
      eventLag: true
    },
    start: function() {
      this.self = Pace.start(this.options);
      $("#envdb-query, #content").css("opacity", 0.5);
      // $("#loading").show();
    },
    stop: function() {
      Pace.stop();
      $("#envdb-query, #content").css("opacity", 1);
      // $("#loading").hide();
    },
    restart: function() {
      Pace.restart();
    },
    done: function() {
      Pace.stop();
      $("#envdb-query, #content").css("opacity", 1);
      // $("#loading").hide();
    }
  },

  Templates: {

    Init: function() {
      this.table = Handlebars.compile($("#query-results-table").html());
      this.row = Handlebars.compile($("#query-results-row").html());
      this.agent = Handlebars.compile($("#agent-template").html());
    }

  },

  Query: {
    Execute: function() {

      $("#content").scrollTop(0);

      if (Envdb.table) {
        Envdb.table.destroy();
        Envdb.table = false;
      }

      if (Envdb.fixedTable) {
        $(".FixedHeader_Cloned").remove();
        Envdb.fixedTable = false;
      }

      var value = Envdb.Editor.self.getValue();

      Envdb.Loading.start()

      Envdb.Query.Run("query", value.replace(/(\r\n|\n|\r)/gm, " "), function(results, err) {

        var table = null;

        if (results && results.length > 0) {

          for (record in results) {

            var agent = results[record];

            agent.results = JSON.parse(agent.results)

            if (!table) {
              var data = {
                name: agent.name,
                hostname: agent.hostname,
                results: agent.results[0]
              }
              table = Envdb.Templates.table(data);
              $("#content .wrapper").html(table);
            }

            var data = {
              name: agent.name,
              hostname: agent.hostname,
              results: agent.results
            }
            var row = Envdb.Templates.row(data)
            $("table.query-results tbody").append(row);

          }

          Envdb.table = $("table.query-results")
            .on('order.dt', function() {
              if (Envdb.fixedTable) {
                Envdb.fixedTable.fnUpdate()
                $("#content").scrollTop(0);
              }
            }).DataTable({
            searching: false,
            paging: false,
            info: false,
            initComplete: function (table){
              Envdb.fixedTable = new $.fn.dataTable.FixedHeader(table);
            }
          });

          window.onresize = function() {
            if (Envdb.fixedTable) {
              Envdb.fixedTable.fnUpdate()
            }
          }

        } else {
          // error - no data
        }

        // $("table.query-results").tablesorter();

        Envdb.Editor.self.focus();
        Envdb.Loading.done()
      });

    },
    Run: function(type, sql, callback) {

      Envdb.Socket.request(type, {
        id: "all",
        sql: sql,
      }, function(err, data) {

          if (typeof callback === "function") {
            return callback(data, err);
          }

        });
    }
  },

  Editor: {
    self: null,

    Build: function() {
      ace.require("ace/ext/language_tools");

      this.self = ace.edit("editor");

      this.self.setOptions({
        enableBasicAutocompletion: true
      });

      this.self.getSession().setMode("ace/mode/sql");

      this.self.getSession().setTabSize(2);
      this.self.getSession().setUseSoftTabs(true);
      this.self.getSession().setUseWrapMode(true);
      this.self.setShowPrintMargin(false);
      this.self.setValue("select * from listening_ports a join processes b on a.pid = b.pid;");

      this.self.focus();
      // this.self.setHighlightActiveLine(false);

      // document.getElementById('editor').style.fontSize='13px';

      this.self.commands.addCommands([
        {
          name: "run_query",
          bindKey: {
            win: "Ctrl-Enter",
            mac: "Command-Enter"
          },
          exec: function(editor) {
            Envdb.Query.Execute();
          }
        }
      ]);

      $("a.run-query").on("click", function(e) {
        e.preventDefault()
        Envdb.Query.Execute();
      });
    }
  },

  Socket: null,
  Init: function() {

    gotalk.handleNotification('agent-update', function(agent) {
      var item = $("li[data-agent-id='" + agent.id + "']");
      if (item.length > 0) {
        item.replaceWith(Envdb.Templates.agent(agent))
      } else {
        $("ul#agents").append(Envdb.Templates.agent(agent))
      }
    });

    Envdb.Socket = gotalk.connection().on('open', function() {});

    Envdb.Templates.Init()
    Envdb.Editor.Build()
  }
};

jQuery(document).ready(function($) {

  Envdb.Init();

  // var lastScrollLeft = 0;
  // $("#content").on("scroll", function() {
    // if (Envdb.fixedTable) {
      // var documentScrollLeft = $("#content").scrollLeft();
      // if (lastScrollLeft != documentScrollLeft) {
        // Envdb.fixedTable.fnUpdate();
        // lastScrollLeft = documentScrollLeft;
      // }
    // }
  // })

  var options = {
    valueNames: ['agent-name', 'agent-agent-id']
  };

  var agentList = new List('sidebar', options);

  // $('#search').keyup(function(e) {
  // var self = this;

  // if (e.keyCode == 13) {
  // var value = $(self).val();
  // console.log(value)

  // $("div.results").html("")
  // $("div.results").append("Loading...")

  // Envdb.Socket.request('query', {
  // id: "all",
  // sql: value,
  // }, function(err, data) {

  // console.log("BEFORE ERROR::", data)

  // $("div.results").html("")

  // for (var i = 0, len = data.length; i < len; i++) {
  // var agent = data[i];
  // agent.results = JSON.parse(agent.results)
  // console.log(agent.results)

  // $("div.results").append("<br /><br />")
  // $("div.results").append("<strong>Agent: " + agent.name + " :: " + agent.id + "</strong>")
  // $("div.results").append("<br /><br />")
  // $("div.results").append(JSON.stringify(agent))
  // $("div.results").append("<br /><br />")

  // }
  // });
  // }
  // });

  // $('#search2').keyup(function(e) {
  // var self = this;

  // if (e.keyCode == 13) {
  // var value = $(self).val();
  // console.log(value)

  // $("div.results").html("")
  // $("div.results").append("Loading...")

  // Envdb.Socket.request('tables', {
  // id: value
  // }, function(err, data) {

  // console.log("word?")
  // console.log(err, data)

  // $("div.results").html(data.results)

  // });
  // }
  // });

  // $('#search3').keyup(function(e) {
  // var self = this;

  // if (e.keyCode == 13) {
  // var value = $(self).val();
  // var id = $("#search2").val();
  // console.log(value)

  // $("div.results").html("")
  // $("div.results").append("Loading...")

  // Envdb.Socket.request('table-info', {
  // id: id,
  // sql: "pragma table_info(" + value + ");",
  // }, function(err, data) {

  // console.log("word?")
  // console.log(err, data)

  // $("div.results").html(JSON.stringify(data))

  // });
  // }
  // });

});
