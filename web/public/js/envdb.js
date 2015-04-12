var Envdb = {

  Request: {},

  table: false,
  fixedTable: false,

  lbox: {
    open: function(template, data, args) {
      var self = this;

      if (args && args.hasOwnProperty('width')) {
        if (args.width) {
          Envdb.lbox.options.style.width = args.width;
        }
      }

      var options = $.extend({}, Envdb.lbox.options, args);
      options.template = template;
      options.templateData = data;

      var box = $.limp(options);

      return box;
    },
    options: {
      cache: false,
      adjustmentSize: 0,
      loading: true,
      alwaysCenter: true,
      animation: "pop",
      shadow: "none",
      round: 0,
      distance: 10,
      overlayClick: true,
      enableEscapeButton: true,
      dataType: 'html',
      centerOnResize: true,
      closeButton: false,
      style: {
        '-webkit-outline': 0,
        color: '#000',
        position: 'fixed',
        border: '1px solid #ededed',
        outline: 0,
        zIndex: 10000001,
        opacity: 0,
        // overflow: 'auto',
        background: '#fff'
      },
      inside: {
        background: '#fff',
        padding: '0',
        display: 'block',
        border: 'none',
        overflow: 'visible'
      },
      overlay: {
        background: '#000',
        opacity: 0.7
      },
      onOpen: function() {
      },
      afterOpen: function() {
      },
      afterDestroy: function() {
      },
      onTemplate: function(template, data, limp) {
        return template;
      }
    }
  },

  Node: {
    current: null,
    tables: [],

    commands: {
      confirm: function(data, callback) {
        var box = Envdb.lbox.open(Envdb.Templates.confirm(data), {}, {
          width: "760px",
          disableDefaultAction: true,
          afterClose: function() {
          },
          afterOpen: function() {
          },
          onAction: function() {
            callback();
          }
        });

        box.open();
      },
      disconnect: function(data) {
        this.confirm(data, function() {
          Envdb.Socket.request('disconnect', data.id, function(err) {
            if (err) {
              Envdb.Flash.error(err);
            }

            $.limpClose();
          });
        });
      },
      delete: function(data) {
        this.confirm(data, function() {
          Envdb.Socket.request('delete', data.id, function(err) {
            if (err) {
              Envdb.Flash.error(err);
            } else {
              console.log(data.id)
              $("li.node[data-node-id='"+data.id+"']").remove();
            }

            $.limpClose();
          });
        });
      }
    },

    fetchTables: function(callback) {
      var self = this;

      Envdb.Socket.request('tables', {
        id: self.current
      }, function(err, data) {

        self.tables = data;

        if (typeof callback === "function") {
          return callback(data, err);
        }
      });
    },

    fetchTableInfo: function(table, callback) {
      var self = this;

      if (Envdb.fixedTable) {
        Envdb.fixedTable._fnDestroy();
        Envdb.fixedTable = false;
      }

      if (Envdb.table) {
        Envdb.table.destroy();
        Envdb.table = false;
      }

      Envdb.Loading.start();

      Envdb.Socket.request('table-info', {
        id: self.current,
        sql: "pragma table_info(" + table + ");",
      }, function(err, data) {

          if (typeof callback === "function") {
            data.hideNode = true;
            Envdb.Query.Render([data], err, function() {
              return callback(data, err);
            });
          }

        });
    },

    clean: function() {
      this.current = null;
      this.tables = [];

      if (Envdb.fixedTable) {
        Envdb.fixedTable._fnDestroy();
        Envdb.fixedTable = false;
      }

      if (Envdb.table) {
        Envdb.table.destroy();
        $(".query-results").remove();
        Envdb.table = false;
      }

      $("#content").removeClass("node-view");
      $("#node-tables").remove();
      // $(".save-query, .load-query").addClass("disabled");
      $('.table-filter').addClass("node-view");
      $(".table-filter").show().find("input").val("");
      $("#content").css("left", 460);
    },

    close: function() {
      this.clean();
      $("#header .title").text("Query All Nodes");
      $("a.run-query").text("Run Query");
      // $(".save-query, .load-query").removeClass("disabled");
      $('.table-filter').removeClass("node-view");
      $(".table-filter").hide().find("input").val("");
      $("#content").css("left", 230);
    },

    open: function(name, id) {
      var self = this;

      this.clean();
      this.current = id;

      Envdb.Loading.start();
      this.fetchTables(function(data, err) {
        Envdb.Loading.done();

        $("a.run-query").text("Query Node");
        $("#header .title").text("Query Node: " + data.name + " (" + data.hostname + " / " + data.id + ")");
        $("#content").addClass("node-view");
        $("#wrapper").append(Envdb.Templates.tables(data.results));

        $("ul.tables li").on("click", function(e) {
          e.preventDefault();

          $("ul.tables li").removeClass("selected");
          $(this).addClass("selected");

          var table = $(this).attr("data-table-name");
          self.fetchTableInfo(table, function() {})
        });

        $("ul.tables li:first-child").click();
      });

    }
  },

  Flash: {
    delay: 1500,

    show: function(data, type) {
      var self = this;

      if (self.timeout) {
        clearTimeout(self.timeout);
        self.timeout = null
        self.hide();
      }

      self.hide();

      $("#flash-message").attr("class", type).text(data).show();

      self.timeout = setTimeout(function() {
        $("#flash-message").stop().fadeOut("slow", function() {
          self.hide();
        });
      }, this.delay);
    },
    hide: function() {
      $("#flash-message").attr("class", "").text("").hide();
    },
    error: function(message) {
      this.show(message, "error");
    },
    success: function(message) {
      this.show(message, "success");
    }
  },

  Loading: {
    options: {
      ajax: false,
      document: false,
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
      this.node = Handlebars.compile($("#node-template").html());
      this.tables = Handlebars.compile($("#tables-template").html());
      this.saveQuery = Handlebars.compile($("#save-query-template").html());
      this.loadQuery = Handlebars.compile($("#load-query-template").html());
      this.nodeContextMenu = Handlebars.compile($("#node-context-menu-template").html());
      this.confirm = Handlebars.compile($("#confirm-template").html());
    }

  },

  Query: {

    RunSavedQuery: function(name, query, callback) {
      $("#content").scrollTop(0);

      $("#header .title").text("Saved Query: " + name);
      Envdb.Editor.self.setValue(query);

      if (Envdb.fixedTable) {
        Envdb.fixedTable._fnDestroy();
        Envdb.fixedTable = false;
      }

      if (Envdb.table) {
        Envdb.table.destroy();
        Envdb.table = false;
      }

      Envdb.Loading.start()

      Envdb.Query.Run("query", query.replace(/(\r\n|\n|\r)/gm, " "), function(results, err) {
        Envdb.Query.Render(results, err);
        if (typeof callback === "function") {
          callback();
        }
      });
    },

    Save: function(params) {
      var editor;

      params.query = Envdb.Editor.self.getValue();

      var box = Envdb.lbox.open(Envdb.Templates.saveQuery(params), {}, {
        width: "760px",
        disableDefaultAction: true,
        afterClose: function() {
          editor.destroy();
        },
        afterOpen: function() {
          editor = Envdb.Editor._build("save-query-editor");
          editor.setValue(params.query);
          $("#save-query-name").focus()
        },
        onAction: function() {

          if (Envdb.Request.save) {
            Envdb.Request.save.abort()
          }

          var saveParams = {
            name: $("#save-query-name").val(),
            query: editor.getValue(),
            type: "all"
          }

          if (saveParams.name.length <= 0) {
            $("#save-query-name").addClass("error");
            return
          } else {
            $("#save-query-name").removeClass("error");
          }

          if (saveParams.query.length <= 0) {
            $("#save-query-editor").addClass("error");
            return;
          } else {
            $("#save-query-editor").removeClass("error");
          }

          Envdb.Request.save = $.ajax({
            url: "/query/save",
            type: "post",
            dataType: "json",
            data: saveParams,
            success: function(data) {
              $.limpClose();

              if (data.error && data.error.length > 0) {
                Envdb.Flash.error(data.error);
              } else {
                Envdb.Flash.success("Query saved successfully.");
              }

            },
            error: function(a,b,c) {
              // console.log(a,b,c)
            }
          })
        }
      });

      box.open();
    },

    Load: function(params) {

      if (Envdb.Request.load) {
        Envdb.Request.load.abort();
      }

      Envdb.Request.load = $.ajax({
        url: "/api/v1/queries",
        dataType: "json",
        error: function(a,b,c) {
          // console.log(a,b,c)
        },
        success: function(data) {
          Envdb.Request.load = null;

          var box = Envdb.lbox.open(Envdb.Templates.loadQuery(data), {}, {
            width: "800px",
            disableDefaultAction: true,
            afterClose: function() {},
            afterOpen: function() {
              var queryList = new List('load-query-select', {
                valueNames: ['name', 'query']
              });

              $("a.load-saved-query").on("click", function(e) {
                e.preventDefault();
                var item = $(this).parents("li");
                var query = item.find(".query").text();
                var name = item.find(".name").text();

                Envdb.Query.RunSavedQuery(name, query);
                $.limpClose();
              });

              $("a.delete-saved-query").on("click", function(e) {
                e.preventDefault();

                var item = $(this).parents("li");
                var id = item.attr("data-query-id");

                console.log(id)

                $.ajax({
                  url: "/query/delete",
                  type: "POST",
                  dataType: "json",
                  data: {
                    id: id
                  },
                  success: function(data) {
                    console.log(data)
                    item.fadeOut("fast");
                  }
                });
              });
            },
            onAction: function() {}
          });

          box.open();
        }
      });

    },

    Execute: function() {
      $("#content").scrollTop(0);
      $(".table-filter").show().find("input").val("");

      if (Envdb.fixedTable) {
        Envdb.fixedTable._fnDestroy();
        Envdb.fixedTable = false;
      }

      if (Envdb.table) {
        Envdb.table.destroy();
        Envdb.table = false;
      }

      var value = Envdb.Editor.self.getValue();

      Envdb.Loading.start()

      Envdb.Query.Run("query", value.replace(/(\r\n|\n|\r)/gm, " "), function(results, err) {
        Envdb.Query.Render(results, err);
      });

    },

    Render: function(results, err, callback) {

      if (results && results.length > 0) {
        if (results[0].error.length > 0) {
          var er = results[0].error;
          if (er === "exit status 1") {
            Envdb.Flash.error("Query Syntax Error - Check your query and try again.");
          } else {
            Envdb.Flash.error("Query Error: " + er);
          }
          Envdb.Editor.self.focus();
          Envdb.Loading.done()
          return;
        }
      } else {
        Envdb.Flash.error("Notice: Your query returned no data.")

        Envdb.Editor.self.focus();
        Envdb.Loading.done()
      }

      var table = null;
      var count = 0

      if (results && results.length > 0) {

        for (record in results) {

          var node = results[record];

          node.results = JSON.parse(node.results)

          if (node.results.length <= 0) {
            continue;
          }

          if (!table) {
            var data = {
              hideNode: node.hideNode || false,
              name: node.name,
              hostname: node.hostname,
              results: node.results[0]
            }
            table = Envdb.Templates.table(data);
            $("#content .wrapper").html(table);
          }

          var data = {
            hideNode: node.hideNode || false,
            name: node.name,
            hostname: node.hostname,
            results: node.results
          }
          var row = Envdb.Templates.row(data)
          $("table.query-results tbody").append(row);

          count++;
        }

        if (count <= 0) {
          if (Envdb.fixedTable) {
            Envdb.fixedTable._fnDestroy();
            Envdb.fixedTable = false;
          }

          if (Envdb.table) {
            Envdb.table.destroy();
            Envdb.table = false;
          }

          Envdb.Loading.done()
          Envdb.Flash.error("No results found.");

          $("#content .wrapper").html("");
          return
        }

        Envdb.table = $("table.query-results")
          .on('order.dt', function() {
            if (Envdb.fixedTable) {
              Envdb.fixedTable.fnUpdate()
              $("#content").scrollTop(0);
            }
          }).DataTable({
          searching: true,
          paging: false,
          info: false
        });


        Envdb.fixedTable = new $.fn.dataTable.FixedHeader(Envdb.table, {
        });

        window.onresize = function() {
          if (Envdb.fixedTable) {
            Envdb.fixedTable.fnUpdate()
          }
        }

        if (typeof callback === "function") {
          callback(results, err);
        }

      } else {
        Envdb.Flash.error("Your query returned no data.")
        // error - no data
      }

      // $("table.query-results").tablesorter();

      Envdb.Editor.self.focus();
      Envdb.Loading.done()
    },

    Run: function(type, sql, callback) {

      var id = "all";

      if (Envdb.Node.current) {
        id = Envdb.Node.current;
      }

      Envdb.Socket.request(type, {
        id: id,
        sql: sql,
      }, function(err, data) {

        if (!data || data.length <= 0) {
          return callback(null, "No data");
        }

        for (var i = 0, len = data.length; i < len; i++) {
          if (id !== "all" && data[i]) {
            data[i].hideNode = true
          }
        }

        if (typeof callback === "function") {
          return callback(data, err);
        }

      });
    }
  },

  Editor: {
    self: null,

    _build: function(div) {
      ace.require("ace/ext/language_tools");

      var editor = ace.edit(div);

      editor.setOptions({
        enableBasicAutocompletion: true,
        enableSnippets: true,
        enableLiveAutocompletion: true
      });

      editor.getSession().setMode("ace/mode/sql");
      editor.getSession().setTabSize(2);
      editor.getSession().setUseSoftTabs(true);
      editor.getSession().setUseWrapMode(true);
      editor.setHighlightActiveLine(false);
      editor.setShowPrintMargin(false);
      // document.getElementById('editor').style.fontSize='13px';

      return editor;
    },

    Build: function() {

      this.self = this._build("editor");

      this.self.focus();
      this.self.setValue("select * from listening_ports a join processes b on a.pid = b.pid;");

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
        e.preventDefault();
        Envdb.Query.Execute();
      });

      $("a.export-results").on("click", function(e) {
        e.preventDefault();
        var csv = $("table.query-results").table2CSV({
          delivery: 'value'
        });
        window.location.href = 'data:text/csv;charset=UTF-8,'
        + encodeURIComponent(csv);
      });
    }
  },

  Socket: null,
  Init: function() {

    gotalk.handleNotification('node-update', function(node) {
      var item = $("li[data-node-id='" + node.id + "']");
      if (item.length > 0) {
        item.replaceWith(Envdb.Templates.node(node))
      } else {
        $("ul#nodes").append(Envdb.Templates.node(node))
      }
    });

    Envdb.Socket = gotalk.connection().on('open', function() {});

    Envdb.Templates.Init()
    Envdb.Editor.Build()
  }
};

jQuery(document).ready(function($) {

  Envdb.Init();

  var lastScrollLeft = 0;
  $("#content").on("scroll", function() {
    if (Envdb.fixedTable) {
      var documentScrollLeft = $("#content").scrollLeft();
      if (lastScrollLeft != documentScrollLeft) {
        // super hack
        Envdb.fixedTable.fnPosition();
        $(".FixedHeader_Cloned").css("top", 230);
        lastScrollLeft = documentScrollLeft;
      }
    }
  });

  $(document).on("click", "li.node", function(e) {
    e.preventDefault();

    var name = $(this).find("span.node-name").text();
    var id = $(this).attr("data-node-id");


    if ($(this).hasClass("online")) {
      if ($(this).hasClass("current")) {
        $("li.node").removeClass("current");
        Envdb.Node.close();
      } else {
        Envdb.Node.open(name, id);
        $("li.node").removeClass("current");
        $(this).addClass("current");
      }

    } else {
      Envdb.Flash.error("Node (" + name + ") is current offline.");
    }
  });

  $(document).on("contextmenu", "li.node", function(event) { 
    event.preventDefault();
    var id = $(this).attr("data-node-id");
    var menu = $(Envdb.Templates.nodeContextMenu({
      id: id
    })).appendTo("body")
    .css({top: event.pageY + "px", left: event.pageX + "px"});

    $("a.disconnect-node").on("click", function(e) {
      e.preventDefault();
      Envdb.Node.commands.disconnect({
        id: id,
        title: "Disconnect Node",
        message: "Are you sure you want to disconnect this node?",
        icon: "fa-exclamation",
        buttonTitle: "Disconnect"
      });
    });

    $("a.delete-node").on("click", function(e) {
      e.preventDefault();
      Envdb.Node.commands.delete({
        id: id,
        title: "Delete Node",
        message: "Are you sure you want to delete this node?",
        icon: "fa-trash",
        buttonTitle: "Delete"
      });
    });

  }).on("click", function(event) {
    $("ul.custom-menu").hide();
  });


  var nodeList = new List('sidebar', {
    valueNames: ['node-name', 'node-node-id']
  });

  $(document).on("click", ".envdb-control a.save-query", function(e) {
    e.preventDefault();
    Envdb.Query.Save({});
  });

  $(document).on("click", ".envdb-control a.load-query", function(e) {
    e.preventDefault();
    Envdb.Query.Load({});
  });

  $(document).on("input", ".table-filter input", function(e) {
    var value = $(this).val();
    // search
    if (Envdb.table) {
      Envdb.table.search(value);
      Envdb.table.draw();
      Envdb.fixedTable.fnUpdate();
    }
  });

  $(document).on("click", "a.logout", function(e) {
    e.preventDefault();
    $.ajax({
      url: "/login",
      type: "DELETE",
      success: function() {
        window.location = "/login";
      }
    })
  });

  $(document).on("click", "i.hide-sidebar", function(e) {
    e.preventDefault();
    $("#sidebar").css("left", -230)
    $("#header, #envdb-query, #node-tables").css("left", 0);

    if ($("#content").hasClass("node-view")) {
      $("#content").css("left", 230);
    } else {
      $("#content").css("left", 0);
    }

    $("div.show-sidebar").show();

    if (Envdb.table) {
      Envdb.table.draw();

      if (Envdb.fixedTable) {
        Envdb.fixedTable.fnUpdate()
        Envdb.fixedTable.fnPosition();
      }
    }
  });

  $(document).on("click", "div.show-sidebar", function(e) {
    e.preventDefault();
    $("#sidebar").css("left", 0)
    $("#header, #envdb-query, #node-tables").css("left", 230);
    $("div.show-sidebar").hide();

    if ($("#content").hasClass("node-view")) {
      $("#content").css("left", 460);
    } else {
      $("#content").css("left", 230);
    }

    if (Envdb.table) {
      Envdb.table.draw();

      if (Envdb.fixedTable) {
        Envdb.fixedTable.fnUpdate()
        Envdb.fixedTable.fnPosition();
      }
    }
  });

});
