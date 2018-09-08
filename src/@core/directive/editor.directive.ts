import { Directive, EventEmitter, Output, ElementRef, Input, OnInit, OnDestroy, NgZone } from "@angular/core";
import * as ace from "brace";
import "brace/theme/chrome";
import "brace/theme/dracula";
import "brace/mode/abap";
import "brace/mode/abc";
import "brace/mode/actionscript";
import "brace/mode/ada";
import "brace/mode/apache_conf";
import "brace/mode/applescript";
import "brace/mode/asciidoc";
import "brace/mode/assembly_x86";
import "brace/mode/autohotkey";
import "brace/mode/batchfile";
import "brace/mode/bro";
import "brace/mode/c9search";
import "brace/mode/c_cpp";
import "brace/mode/cirru";
import "brace/mode/clojure";
import "brace/mode/cobol";
import "brace/mode/coffee";
import "brace/mode/coldfusion";
import "brace/mode/csharp";
import "brace/mode/csound_document";
import "brace/mode/csound_orchestra";
import "brace/mode/csound_score";
import "brace/mode/css";
import "brace/mode/curly";
import "brace/mode/d";
import "brace/mode/dart";
import "brace/mode/diff";
import "brace/mode/django";
import "brace/mode/dockerfile";
import "brace/mode/dot";
import "brace/mode/drools";
import "brace/mode/eiffel";
import "brace/mode/ejs";
import "brace/mode/elixir";
import "brace/mode/elm";
import "brace/mode/erlang";
import "brace/mode/forth";
import "brace/mode/fortran";
import "brace/mode/ftl";
import "brace/mode/gcode";
import "brace/mode/gherkin";
import "brace/mode/gitignore";
import "brace/mode/glsl";
import "brace/mode/gobstones";
import "brace/mode/golang";
import "brace/mode/graphqlschema";
import "brace/mode/groovy";
import "brace/mode/haml";
import "brace/mode/handlebars";
import "brace/mode/haskell";
import "brace/mode/haskell_cabal";
import "brace/mode/haxe";
import "brace/mode/hjson";
import "brace/mode/html";
import "brace/mode/html_elixir";
import "brace/mode/html_ruby";
import "brace/mode/ini";
import "brace/mode/io";
import "brace/mode/jack";
import "brace/mode/jade";
import "brace/mode/java";
import "brace/mode/javascript";
import "brace/mode/json";
import "brace/mode/jsoniq";
import "brace/mode/jsp";
import "brace/mode/jssm";
import "brace/mode/jsx";
import "brace/mode/julia";
import "brace/mode/kotlin";
import "brace/mode/latex";
import "brace/mode/lean";
import "brace/mode/less";
import "brace/mode/liquid";
import "brace/mode/lisp";
import "brace/mode/live_script";
import "brace/mode/livescript";
import "brace/mode/logiql";
import "brace/mode/lsl";
import "brace/mode/lua";
import "brace/mode/luapage";
import "brace/mode/lucene";
import "brace/mode/makefile";
import "brace/mode/markdown";
import "brace/mode/mask";
import "brace/mode/matlab";
import "brace/mode/mavens_mate_log";
import "brace/mode/maze";
import "brace/mode/mel";
import "brace/mode/mips_assembler";
import "brace/mode/mipsassembler";
import "brace/mode/mushcode";
import "brace/mode/mysql";
import "brace/mode/nix";
import "brace/mode/nsis";
import "brace/mode/objectivec";
import "brace/mode/ocaml";
import "brace/mode/pascal";
import "brace/mode/perl";
import "brace/mode/pgsql";
import "brace/mode/php";
import "brace/mode/pig";
import "brace/mode/plain_text";
import "brace/mode/powershell";
import "brace/mode/praat";
import "brace/mode/prolog";
import "brace/mode/properties";
import "brace/mode/protobuf";
import "brace/mode/python";
import "brace/mode/r";
import "brace/mode/razor";
import "brace/mode/rdoc";
import "brace/mode/red";
import "brace/mode/rhtml";
import "brace/mode/rst";
import "brace/mode/ruby";
import "brace/mode/rust";
import "brace/mode/sass";
import "brace/mode/scad";
import "brace/mode/scala";
import "brace/mode/scheme";
import "brace/mode/scss";
import "brace/mode/sh";
import "brace/mode/sjs";
import "brace/mode/smarty";
import "brace/mode/snippets";
import "brace/mode/soy_template";
import "brace/mode/space";
import "brace/mode/sparql";
import "brace/mode/sql";
import "brace/mode/sqlserver";
import "brace/mode/stylus";
import "brace/mode/svg";
import "brace/mode/swift";
import "brace/mode/swig";
import "brace/mode/tcl";
import "brace/mode/tex";
import "brace/mode/text";
import "brace/mode/textile";
import "brace/mode/toml";
import "brace/mode/tsx";
import "brace/mode/turtle";
import "brace/mode/twig";
import "brace/mode/typescript";
import "brace/mode/vala";
import "brace/mode/vbscript";
import "brace/mode/velocity";
import "brace/mode/verilog";
import "brace/mode/vhdl";
import "brace/mode/wollok";
import "brace/mode/xml";
import "brace/mode/xquery";
import "brace/mode/yaml";

/**
 * EditorDirective is a port of Ace editor (brace version) for Angular framework.
 *
 * @export
 * @class EditorDirective
 * @implements {OnInit}
 * @implements {OnDestroy}
 */
@Directive({
    selector: 'editor, [editor]'
})
export class EditorDirective implements OnInit, OnDestroy {

    @Output() onChange = new EventEmitter();
    @Output() onChanged = new EventEmitter();

    private _editor: any;
    private _options: any = {};
    private _readOnly: boolean = false;
    private _theme: string = "chrome";
    private _mode: any = "text";
    private _autoUpdateContent: boolean = true;
    private _durationBeforeCallback: number = 0;
    private _text: string = "";

    private oldText: any;
    private timeoutSaving: any;

    constructor(elementRef: ElementRef, private zone: NgZone) {
        let el = elementRef.nativeElement;
        this.zone.runOutsideAngular(() => {
            this._editor = ace.edit(el);
        });
        this._editor.$blockScrolling = Infinity;
    }

    ngOnInit() {
        this._editor.setOptions(this._options || {});
        this._editor.setTheme(`ace/theme/${this._theme}`);
        this.setMode(this._mode);
        this._editor.setReadOnly(this._readOnly);
        
        this._editor.on('change', () => this.updateText());
        this._editor.on('paste', () => this.updateText());
    }

    ngOnDestroy() {
        this._editor.destroy();
    }

    private updateText() {
        let newVal = this._editor.getValue();
        if (newVal === this.oldText) {
            return;
        }
        if (!this._durationBeforeCallback) {
            this._text = newVal;
            this.zone.run(() => {
                this.onChange.emit(newVal);
                this.onChanged.emit(newVal);
            });
        } else {
            if (this.timeoutSaving != null) {
                clearTimeout(this.timeoutSaving);
            }

            this.timeoutSaving = setTimeout(() => {
                this._text = newVal;
                this.zone.run(() => {
                    this.onChange.emit(newVal);
                    this.onChanged.emit(newVal);
                });
                this.timeoutSaving = null;
            }, this._durationBeforeCallback);
        }
        this.oldText = newVal;
    }

    @Input() set options(options: any) {
        this._options = options;
        this._editor.setOptions(options || {});
    }

    @Input() set readOnly(readOnly: any) {
        this._readOnly = readOnly;
        this._editor.setReadOnly(readOnly);
    }

    @Input() set theme(theme: any) {
        this._theme = theme;
        this._editor.setTheme(`ace/theme/${theme}`);
    }

    @Input() set mode(mode: any) {
        this.setMode(mode);
    }

    private setMode(mode: any) {
        this._mode = mode;
        if (typeof this._mode === 'object') {
            this._editor.getSession().setMode(this._mode);
        } else {
            this._editor.getSession().setMode(`ace/mode/${this._mode}`);
        }
    }

    @Input()
    get text() {
        return this._text;
    }
    set text(text: string) {
        this.setText(text);
    }

    private setText(text: any) {
        if (this._text !== text) {
            if (text === null || text === undefined) {
                text = "";
            }

            if (this._autoUpdateContent === true) {
                this._text = text;
                this._editor.setValue(text);
                this._editor.clearSelection();
            }
        }
    }

    @Input() set autoUpdateContent(status: any) {
        this._autoUpdateContent = status;
    }

    @Input() set durationBeforeCallback(num: number) {
        this._durationBeforeCallback = num;
    }

    get editor(): any {
        return this._editor;
    }
}