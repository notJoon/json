package json

const (
	SquareOpenToken     = '['
	SquareCloseToken    = ']'
	RoundOpenToken      = '('
	RoundCloseToken     = ')'
	CurlyOpenToken      = '{'
	CurlyCloseToken     = '}'
	CommaToken          = ','
	DotToken            = '.'
	ColonToken          = ':'
	BackTickToken       = '`'
	QuoteToken          = '\''
	DoublyQuoteToken    = '"'
	EmptyStringToken    = ""
	WhiteSpaceToken     = ' '
	PlusToken           = '+'
	MinusToken          = '-'
	AesteriskToken      = '*'
	BangToken           = '!'
	QuestionToken       = '?'
	NewLineToken        = '\n'
	TabToken            = '\t'
	CarriageReturnToken = '\r'
	FormFeedToken       = '\f'
	BackSpaceToken      = '\b'
	SlashToken          = '/'
	BackSlashToken      = '\\'
	UnderScoreToken     = '_'
	DollarToken         = '$'
	AtToken             = '@'
	AndToken            = '&'
	OrToken             = '|'
)

var (
	trueLiteral  = []byte("true")
	falseLiteral = []byte("false")
	nullLiteral  = []byte("null")
)

type ValueType int

const (
	NotExist ValueType = iota
	String
	Number
	Float
	Object
	Array
	Boolean
	Null
	Unknown
)

func (v ValueType) String() string {
	switch v {
	case NotExist:
		return "not-exist"
	case String:
		return "string"
	case Number:
		return "number"
	case Object:
		return "object"
	case Array:
		return "array"
	case Boolean:
		return "boolean"
	case Null:
		return "null"
	default:
		return "unknown"
	}
}