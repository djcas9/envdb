(function(b) {
    function e(a, c) {
        this.options = c;
        this.$element = b(a);
        this.inProgress = !1;
        this.template = this.options.template || null;
        this.templateData = this.options.templateData || {};
        this.url = this.options.url || null;
        this.cache = this.$loader = this.$limp = !1;
        this.$element.data("template") && (this.template = this.$element.data("template"));
        this.$element.data("width") && (this.options.style.width = this.$element.data("width"));
        this.template || (this.$element.attr("href") ? this.url = this.$element.attr("href") : this.$element.data("url") &&
            (this.url = this.$element.data("url")));
        return this
    }
    e.prototype = {
        open: function() {
            $("*").blur()
            var a = this;
            if (!a.inProgress) return a.inProgress = !0, a.limp(), b(document).bind("limp.close", function() {
                a.close()
            }), a.enableEscapeButtonHandler = function(b) {
                b.stopPropagation();
                27 == b.keyCode && a.close()
            }, a.onActionHandler = function(a) {
                a.stopPropagation();
                13 == a.keyCode && (a = b("button.limp-action")) && a.click()
            }, a.options.enableEscapeButton && b(document).bind("keydown", a.enableEscapeButtonHandler), a.$overlay.prependTo("body"), a.fetch(function(c) {
                a.visible = !0;
                a.options.onOpen(a);
                a.options.overlayClick && a.$overlay.bind("click", function(b) {
                    b.preventDefault();
                    a.close()
                });
                a.options.centerOnResize && b(window).bind("limp.resize", function() {
                    a.visible && a.resize()
                });
                a.$limp.find("#limp-box-inside").html(c);
                a.$limp.css({
                    maxHeight: b(window).height() - 2 * a.options.distance,
                    display: "block"
                }).prependTo("body");
                a.options.afterOpen(a, a.$limp);
                if ("pop" == a.options.animation) {
                    a.$overlay.animate({
                        opacity: a.options.overlay.opacity
                    }, "fast");
                    var d = a.resize(200);
                    a.$limp.animate({
                        top: d.height,
                        opacity: 1
                    }, 400);
                    a.onClose = function(b, c) {
                        a.$overlay.fadeOut("fast");
                        a.$limp.animate({
                            top: d.offsetHeight,
                            opacity: 0
                        }, 400, function() {
                            c()
                        })
                    }
                } else "fade" == a.options.animation ? (a.resize(), a.$limp.animate({
                    opacity: 1
                }, "fast"), a.$overlay.animate({
                    opacity: a.options.overlay.opacity
                }, "fast"), a.onClose = function(b, c) {
                    a.$overlay.fadeOut("fast");
                    b.$limp.fadeOut("fast", function() {
                        c()
                    })
                }) : (a.$overlay.css({
                    opacity: 0.8
                }), a.resize(), a.$limp.css("opacity", 1), a.onClose = function(a, b) {
                    b()
                });
                a.$limp.data("limp-api", a);
                a.options.disableDefaultAction ||
                    b(document).bind("keydown", a.onActionHandler);
                a.options.onAction && "function" === typeof a.options.onAction && b(".limp-action", a.$limp).bind("click", function(b) {
                    b.preventDefault();
                    a.options.onAction()
                })
            }), a
        },
        resize: function(a) {
            var c = 0,
                d = b(window).height() / 2,
                c = b(window).width() / 2;
            "undefined" === typeof this.limpStartSize && (this.limpStartSize = this.$limp.outerHeight(!0));
            this.options.adjustmentSize ? 500 > this.limpStartSize ? 400 < b(window).height() ? (d = d - this.options.adjustmentSize - this.$limp.outerHeight(!0) / 2,
                d <= this.options.adjustmentSize && (d = this.options.adjustmentSize)) : d -= this.$limp.outerHeight(!0) / 2 : d -= this.$limp.outerHeight(!0) / 2 : d -= this.$limp.outerHeight(!0) / 2;
            var f = c - this.$limp.outerWidth(!0) / 2,
                c = a ? d + a : d;
            this.$limp.css({
                maxHeight: b(window).height() - 2 * this.options.distance,
                top: c,
                left: f
            });
            return {
                height: d,
                width: f,
                offset: a,
                offsetHeight: c
            }
        },
        close: function() {
            var a = this;
            a.inProgress = !1;
            a.onClose(a, function() {
                a.clean()
            });
            return a
        },
        clean: function() {
            this.visible = !1;
            this.$limp.css("opacity", 0);
            b(window).unbind("limp.resize");
            b(document).unbind("keydown", this.enableEscapeButtonHandler);
            b(document).unbind("keydown", this.onActionHandler);
            b(document).unbind("limp.close");
            this.options.onOpen(this);
            this.$limp.remove();
            this.$overlay.remove();
            this.options.afterClose(this, this.$limp);
            this.options.afterDestroy(this, this.$limp)
        },
        error: function(a, c, d) {
            a = b('<span class="limp-box-error" />');
            try {
                var f = c.charAt(0).toUpperCase() + c.slice(1),
                    g = a.append("" + f + ": " + d + ".")
            } catch (e) {
                g = a.append("<div class='limp-error-box'><div class='limp-error-box-title'><div class='icon' />Loading Error</div><div class='limp-error-box-content'><span>Error: Not Found</span></div><div class='limp-error-box-footer'><div class='form-actions'><button class='form-button default' onClick='$.limpClose()'>Ok</button></div></div></div>")
            }
            return g
        },
        fetch: function(a) {
            var c = this;
            b.fn.limp.loading && (b.fn.limp.loading.remove(), b.fn.limp.loading = !1);
            c.url ? c.cache && c.options.cache ? a(c.cache) : b.ajax({
                url: c.url,
                type: "GET",
                dataType: c.options.dataType,
                success: function(b) {
                    c.cache = b;
                    a(b)
                },
                error: function(b, f, e) {
                    c.cache = !1;
                    a(c.error(b, f, e))
                }
            }) : c.template ? c.options.onTemplate && "function" === typeof c.options.onTemplate && (c.cache = c.options.onTemplate(c.template, c.templateData, c), c.cache ? a(c.cache) : a(c.error())) : a(c.error());
            return !1
        },
        limp: function() {
            var a = this;
            a.$limp || (a.$overlay = b('<div id="limp-box-overlay" />').css({
                background: a.options.overlay.background,
                opacity: 0,
                position: "fixed",
                display: "block",
                left: 0,
                top: 0,
                right: 0,
                bottom: 0,
                zIndex: 1E6
            }), a.$limp = b('<div id="limp-box" />'), a.$limpInside = b('<div id="limp-box-inside" />'), a.$limpInside.css(a.options.inside).appendTo(a.$limp), a.options.closeButton && (b('<div id="limp-box-close" class="limp-box-close"><div class="limp-box-close-icon" /><i class="fa fa-remove"></i></div>').css({
                position: "absolute",
                display: "block",
                top: 15,
                right: 18,
                height: 22,
                width: 22,
                cursor: "pointer",
                position: "absolute",
                color: "#999",
                "line-height": "10px",
                "font-size": "18px"
            }).appendTo(a.$limp), b("body").on("click", ".limp-box-close", function(a) {
                a.preventDefault();
                b(document).trigger("limp.close")
            })), a.css());
            b(window).resize(function() {
                a.$limp.trigger("limp.resize")
            });
            return a.$limp
        },
        css: function() {
            this.$limp.css(this.options.style);
            this.options.shadow && this.shadow();
            this.options.round && this.round()
        },
        shadow: function() {
            this.$limp.css({
                "-webkit-box-shadow": this.options.shadow,
                "-moz-box-shadow": this.options.shadow,
                "box-shadow": this.options.shadow
            })
        },
        round: function() {
            this.$limp.css({
                "border-radius": this.options.round,
                "-moz-border-radius": this.options.round,
                "-webkit-border-radius": this.options.round
            });
            this.$limp.find("#limp-box-inside").css({
                "border-radius": this.options.round,
                "-moz-border-radius": this.options.round,
                "-webkit-border-radius": this.options.round
            })
        },
        loading: function() {
            b.fn.limp.loading.remove();
            b.fn.limp.loading = !1
        }
    };
    b.limpResize = function() {
        b(document).trigger("limp.resize");
        return !1
    };
    b.limpClose = function() {
        b(document).trigger("limp.close");
        return !1
    };
    b.limp = function(a) {
        a = b.extend({}, b.fn.limp.defaults, a);
        a.style = b.extend({}, b.fn.limp.style, a.style);
        a.overlay = b.extend({}, b.fn.limp.overlay, a.overlay);
        a.inside = b.extend({}, b.fn.limp.inside, a.inside);
        return new e(null, a)
    };
    b.fn.limp = function(a) {
        var c = [];
        a = b.extend({}, b.fn.limp.defaults, a);
        a.style = b.extend({}, b.fn.limp.style, a.style);
        a.overlay = b.extend({}, b.fn.limp.overlay, a.overlay);
        a.inside = b.extend({}, b.fn.limp.inside, a.inside);
        this.live("click", function(d) {
            d.preventDefault();
            d = b.data(this,
                "limp");
            d || (d = new e(this, a), b.data(this, "limp", d), c.push(d));
            d.visible ? d.close() : d.open()
        });
        return this
    };
    b.fn.limp.defaults = {
        cache: !1,
        disableDefaultAction: !1,
        adjustmentSize: null,
        loading: !0,
        alwaysCenter: !0,
        round: 0,
        animation: !1,
        shadow: "0 1px 10px rgba(0,0,0,0.2)",
        distance: 50,
        overlayClick: !0,
        enableEscapeButton: !0,
        dataType: "html",
        centerOnResize: !0,
        closeButton: !0,
        onOpen: function() {},
        afterOpen: function() {},
        onClose: function() {},
        afterDestroy: function() {},
        afterClose: function() {},
        onTemplate: function() {}
    };
    b.fn.limp.style = {
        "-webkit-outline": 0,
        background: "#fff",
        color: "#000",
        position: "fixed",
        width: "700px",
        border: "solid 5px #ededed",
        color: "black",
        outline: 0,
        zIndex: 1000001,
        opacity: 0,
        height: "auto",
        overflow: "visible"
    };
    b.fn.limp.inside = {
        background: "#fff",
        padding: "35px 40px",
        display: "block",
        border: "1px solid #ddd",
        overflow: "visible"
    };
    b.fn.limp.overlay = {
        background: "#fff",
        opacity: 0.4
    };
    b.fn.limp.loading = !1;
    b.limpLoading = function() {
        b.fn.limp.loading ? (b.fn.limp.loading.remove(), b.fn.limp.loading = !1) : (b.fn.limp.loading = b('<div id="limp-box-loading" />'),
            b.fn.limp.loading.css({
                "border-radius": 5,
                "-moz-border-radius": 5,
                "-webkit-border-radius": 5,
                width: "21px",
                height: "16px",
                top: b(window).height() / 2 - 18.5,
                left: b(window).width() / 2 - 18.5,
                display: "block",
                padding: "10px",
                zIndex: 1000002,
                position: "fixed",
                background: "#111111",
                "background-image": "url(data:image/gif;base64,R0lGODlhEAALAPQAABEREf///zIyMjs7OyMjI/j4+P///9PT04SEhKSkpFBQUN7e3ri4uH19faCgoExMTNra2vr6+rW1tScnJzQ0NBoaGsnJyTAwMBwcHFRUVGhoaEFBQR8fHwAAAAAAAAAAACH+GkNyZWF0ZWQgd2l0aCBhamF4bG9hZC5pbmZvACH5BAALAAAAIf8LTkVUU0NBUEUyLjADAQAAACwAAAAAEAALAAAFLSAgjmRpnqSgCuLKAq5AEIM4zDVw03ve27ifDgfkEYe04kDIDC5zrtYKRa2WQgAh+QQACwABACwAAAAAEAALAAAFJGBhGAVgnqhpHIeRvsDawqns0qeN5+y967tYLyicBYE7EYkYAgAh+QQACwACACwAAAAAEAALAAAFNiAgjothLOOIJAkiGgxjpGKiKMkbz7SN6zIawJcDwIK9W/HISxGBzdHTuBNOmcJVCyoUlk7CEAAh+QQACwADACwAAAAAEAALAAAFNSAgjqQIRRFUAo3jNGIkSdHqPI8Tz3V55zuaDacDyIQ+YrBH+hWPzJFzOQQaeavWi7oqnVIhACH5BAALAAQALAAAAAAQAAsAAAUyICCOZGme1rJY5kRRk7hI0mJSVUXJtF3iOl7tltsBZsNfUegjAY3I5sgFY55KqdX1GgIAIfkEAAsABQAsAAAAABAACwAABTcgII5kaZ4kcV2EqLJipmnZhWGXaOOitm2aXQ4g7P2Ct2ER4AMul00kj5g0Al8tADY2y6C+4FIIACH5BAALAAYALAAAAAAQAAsAAAUvICCOZGme5ERRk6iy7qpyHCVStA3gNa/7txxwlwv2isSacYUc+l4tADQGQ1mvpBAAIfkEAAsABwAsAAAAABAACwAABS8gII5kaZ7kRFGTqLLuqnIcJVK0DeA1r/u3HHCXC/aKxJpxhRz6Xi0ANAZDWa+kEAA7AAAAAAAAAAAA)",
                "background-repeat": "no-repeat",
                "background-position": "center center"
            }), b.fn.limp.loading.prependTo("body"))
    }
})(jQuery);
