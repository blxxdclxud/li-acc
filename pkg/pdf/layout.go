package pdf

// Frame is a rectangle that represents some field on pdf receipt.
// In case of current receipts is used to set coordinates of cells that contain
// credentials (in Excel format they are in column 2).
// MarginXxxx fields are used to set the margin for text inside the frame.
type Frame struct {
	X, Y, W, H float64
	Margin     Margin
}

// InnerRect returns the coordinates of the rectangle inside the frame
// after applying margins. Useful for text placement inside fields.
func (f Frame) InnerRect() (x, y, w, h float64) {
	return f.X + f.Margin.Left,
		f.Y + f.Margin.Top,
		f.W - f.Margin.Left - f.Margin.Right,
		f.H - f.Margin.Top - f.Margin.Bottom
}

// Margin defines padding (in points) inside a Frame.
type Margin struct {
	Top, Right, Bottom, Left float64
}

// Offsets (by Y-coordinate) for certain frames relative to the main frame MainFrame.
// All offsets are selected empirically.
const (
	OffsetPayerCredentialsTopY    = 90
	OffsetPayerCredentialsBottomY = 281

	OffsetPaymentAmountTopY    = 125
	OffsetPaymentAmountBottomY = 316
)

// ----- Margins -----
// Margins are chosen empirically for each group of fields.
var (
	// MarginPayerCredentials is used inside payer credentials frames
	// (top and bottom parts). Small padding on all sides to fit multiline text.
	MarginPayerCredentials = Margin{Top: 14, Right: 4, Bottom: 2, Left: 4}

	// MarginPaymentAmount is used inside payment amount frames
	// (top and bottom parts). Wide left/right padding to center the text.
	MarginPaymentAmount = Margin{Top: 17, Right: 104, Bottom: 2, Left: 104}
)

// ----- FRAMES -----

// Frames contain coordinates and margins of itself. All numbers are selected empirically.
var (
	// MainFrame is frame containing all cells where some data will be shown:
	// cell with organization credentials, cell with payer credentials, etc. These cells are in one column.
	MainFrame = Frame{X: 161, Y: 14, W: 381, H: 383}
)

// Payer credentials section (two cells: top and bottom).
var (
	// FramePayerCredentialsTop contains the top part of payer credentials.
	FramePayerCredentialsTop = Frame{
		X:      MainFrame.X,
		Y:      MainFrame.Y + OffsetPayerCredentialsTopY,
		W:      MainFrame.W,
		H:      36,
		Margin: MarginPayerCredentials}

	// FramePayerCredentialsBottom contains the bottom part of payer credentials.
	FramePayerCredentialsBottom = Frame{
		X:      MainFrame.X,
		Y:      MainFrame.Y + OffsetPayerCredentialsBottomY,
		W:      MainFrame.W,
		H:      36,
		Margin: MarginPayerCredentials}
)

// Payment amount section (two cells: top and bottom).
var (
	// FramePaymentAmountTop contains the top part of the payment amount field.
	FramePaymentAmountTop = Frame{
		X:      MainFrame.X,
		Y:      MainFrame.Y + OffsetPaymentAmountTopY,
		W:      MainFrame.W,
		H:      26,
		Margin: MarginPaymentAmount}

	// FramePaymentAmountBottom contains the bottom part of the payment amount field.
	FramePaymentAmountBottom = Frame{
		X:      MainFrame.X,
		Y:      MainFrame.Y + OffsetPaymentAmountBottomY,
		W:      MainFrame.W,
		H:      26,
		Margin: MarginPaymentAmount}
)

// FrameQrCode defines the square area for placing the QR code in the receipt.
var FrameQrCode = Frame{X: 35, Y: 270, W: 120, H: 120}
