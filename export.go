package tui

import (
	"github.com/stelmakhdigital/stell-tui/component"
	"github.com/stelmakhdigital/stell-tui/diff"
	"github.com/stelmakhdigital/stell-tui/editor"
	"github.com/stelmakhdigital/stell-tui/keys"
	"github.com/stelmakhdigital/stell-tui/overlay"
	"github.com/stelmakhdigital/stell-tui/terminal"
	"github.com/stelmakhdigital/stell-tui/wrap"
)

// Реэкспорт типов и конструкторов подпакетов — единый импорт github.com/stelmakhdigital/stell-tui.

type (
	Component            = component.Component
	Focusable            = component.Focusable
	InputHandler         = component.InputHandler
	Invalidatable        = component.Invalidatable
	Container            = component.Container
	Text                 = component.Text
	Spacer               = component.Spacer
	TruncatedText        = component.TruncatedText
	Box                  = component.Box
	Loader               = component.Loader
	CancellableLoader    = component.CancellableLoader
	SelectList           = component.SelectList
	SettingsList         = component.SettingsList
	SettingsItem         = component.SettingsItem
	Markdown             = component.Markdown
	MarkdownTheme        = component.MarkdownTheme
	Image                = component.Image

	DiffStrategy = diff.DiffStrategy
	DiffEngine   = diff.DiffEngine

	Editor   = editor.Editor
	Input    = editor.Input
	KillRing = editor.KillRing
	Autocomplete  = editor.Autocomplete
	CompleteItem  = editor.CompleteItem
	Completer     = editor.Completer
	StaticCompleter = editor.StaticCompleter

	KeyMap               = keys.KeyMap
	KeyAction            = keys.KeyAction
	KeyDef               = keys.KeyDef
	KeybindingsManager   = keys.KeybindingsManager
	StdinBuffer          = keys.StdinBuffer
	StdinBufferOptions   = keys.StdinBufferOptions

	Terminal             = terminal.Terminal
	ProcessTerminal      = terminal.ProcessTerminal
	TerminalCapabilities = terminal.TerminalCapabilities
	ImageProtocol        = terminal.ImageProtocol
	ImageRenderOptions   = terminal.ImageRenderOptions

	OverlayAnchor  = overlay.OverlayAnchor
	OverlayMargin  = overlay.OverlayMargin
	OverlayOptions = overlay.OverlayOptions
)

const (
	DiffFull   = diff.DiffFull
	DiffPatch  = diff.DiffPatch
	DiffScroll = diff.DiffScroll

	ImageNone  = terminal.ImageNone
	ImageKitty = terminal.ImageKitty
	ImageITerm = terminal.ImageITerm

	OverlayAnchorTop          = overlay.OverlayAnchorTop
	OverlayAnchorCenter       = overlay.OverlayAnchorCenter
	OverlayAnchorBottom       = overlay.OverlayAnchorBottom
	OverlayAnchorTopLeft      = overlay.OverlayAnchorTopLeft
	OverlayAnchorTopRight     = overlay.OverlayAnchorTopRight
	OverlayAnchorBottomLeft   = overlay.OverlayAnchorBottomLeft
	OverlayAnchorBottomRight  = overlay.OverlayAnchorBottomRight
	OverlayAnchorTopCenter    = overlay.OverlayAnchorTopCenter
	OverlayAnchorBottomCenter = overlay.OverlayAnchorBottomCenter
	OverlayAnchorLeftCenter   = overlay.OverlayAnchorLeftCenter
	OverlayAnchorRightCenter  = overlay.OverlayAnchorRightCenter

	CursorMarker = wrap.CursorMarker
)

var (
	NewContainer         = component.NewContainer
	NewSelectList        = component.NewSelectList
	NewSettingsList      = component.NewSettingsList
	NewMarkdown          = component.NewMarkdown
	DefaultMarkdownTheme = component.DefaultMarkdownTheme
	NewImage             = component.NewImage

	NewDiffEngine = diff.NewDiffEngine

	NewEditor   = editor.NewEditor
	NewInput    = editor.NewInput
	NewKillRing = editor.NewKillRing

	NewKeyMap             = keys.NewKeyMap
	NewKeybindingsManager = keys.NewKeybindingsManager
	DefaultTUIKeybindings = keys.DefaultTUIKeybindings
	ParseKey              = keys.ParseKey
	MatchesKey            = keys.MatchesKey
	NormalizeKeyChord     = keys.NormalizeKeyChord
	NormalizeKey          = keys.NormalizeKey
	DecodePrintableKey    = keys.DecodePrintableKey
	IsKeyRelease          = keys.IsKeyRelease
	NewStdinBuffer        = keys.NewStdinBuffer
	SetKittyProtocolActive = keys.SetKittyProtocolActive
	KittyProtocolActive    = keys.KittyProtocolActive

	NewProcessTerminal           = terminal.NewProcessTerminal
	DetectCapabilities           = terminal.DetectCapabilities
	EnableRawMode                = terminal.EnableRawMode
	EnableTerminalFeatures       = terminal.EnableTerminalFeatures
	EnableTerminalFeaturesWriter = terminal.EnableTerminalFeaturesWriter
	TermSize                     = terminal.TermSize
	QueryCellSize                = terminal.QueryCellSize
	QueryCellSizeWriter          = terminal.QueryCellSizeWriter
	WatchResize                  = terminal.WatchResize
	EncodeTerminalImage          = terminal.EncodeTerminalImage
	ImageStub                    = terminal.ImageStub

	ClampOverlayLines     = overlay.ClampOverlayLines
	CompositeOverlayLines = overlay.CompositeOverlayLines

	VisibleLen  = wrap.VisibleLen
	Truncate    = wrap.Truncate
	FuzzyFilter = wrap.FuzzyFilter
	FuzzyScore  = wrap.FuzzyScore
)
