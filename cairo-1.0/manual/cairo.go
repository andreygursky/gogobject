// Full cairo bindings. Made in a stupid fashion (i.e. no sugar). But bits of
// memory management are here (GC finalizer hooks).

package cairo

/*
#include <stdlib.h>
#include <cairo.h>

extern cairo_status_t io_reader_wrapper(void*, unsigned char*, unsigned int);
static cairo_surface_t * _cairo_image_surface_create_from_png_stream(void *closure)
{
	return cairo_image_surface_create_from_png_stream(io_reader_wrapper, closure);
}

extern cairo_status_t io_writer_wrapper(void*, const unsigned char*, unsigned int);
static cairo_status_t _cairo_surface_write_to_png_stream(cairo_surface_t *surface, void *closure)
{
	return cairo_surface_write_to_png_stream(surface, io_writer_wrapper, closure);
}

#cgo pkg-config: cairo
*/
import "C"
import "runtime"
import "reflect"
import "unsafe"
import "io"

//----------------------------------------------------------------------------
// TODO MOVE
//----------------------------------------------------------------------------

var go_repr_cookie C.cairo_user_data_key_t

type RectangleInt struct {
	X      int32
	Y      int32
	Width  int32
	Height int32
}

func (this *RectangleInt) c() *C.cairo_rectangle_int_t {
	return (*C.cairo_rectangle_int_t)(unsafe.Pointer(this))
}

//----------------------------------------------------------------------------
// Error Handling
//----------------------------------------------------------------------------

type Status int

const (
	StatusSuccess                Status = C.CAIRO_STATUS_SUCCESS
	StatusNoMemory               Status = C.CAIRO_STATUS_NO_MEMORY
	StatusInvalidRestore         Status = C.CAIRO_STATUS_INVALID_RESTORE
	StatusInvalidPopGroup        Status = C.CAIRO_STATUS_INVALID_POP_GROUP
	StatusNoCurrentPoint         Status = C.CAIRO_STATUS_NO_CURRENT_POINT
	StatusInvalidMatrix          Status = C.CAIRO_STATUS_INVALID_MATRIX
	StatusInvalidStatus          Status = C.CAIRO_STATUS_INVALID_STATUS
	StatusNullPointer            Status = C.CAIRO_STATUS_NULL_POINTER
	StatusInvalidString          Status = C.CAIRO_STATUS_INVALID_STRING
	StatusInvalidPathData        Status = C.CAIRO_STATUS_INVALID_PATH_DATA
	StatusReadError              Status = C.CAIRO_STATUS_READ_ERROR
	StatusWriteError             Status = C.CAIRO_STATUS_WRITE_ERROR
	StatusSurfaceFinished        Status = C.CAIRO_STATUS_SURFACE_FINISHED
	StatusSurfaceTypeMismatch    Status = C.CAIRO_STATUS_SURFACE_TYPE_MISMATCH
	StatusPatternTypeMismatch    Status = C.CAIRO_STATUS_PATTERN_TYPE_MISMATCH
	StatusInvalidContent         Status = C.CAIRO_STATUS_INVALID_CONTENT
	StatusInvalidFormat          Status = C.CAIRO_STATUS_INVALID_FORMAT
	StatusInvalidVisual          Status = C.CAIRO_STATUS_INVALID_VISUAL
	StatusFileNotFound           Status = C.CAIRO_STATUS_FILE_NOT_FOUND
	StatusInvalidDash            Status = C.CAIRO_STATUS_INVALID_DASH
	StatusInvalidDscComment      Status = C.CAIRO_STATUS_INVALID_DSC_COMMENT
	StatusInvalidIndex           Status = C.CAIRO_STATUS_INVALID_INDEX
	StatusClipNotRepresentable   Status = C.CAIRO_STATUS_CLIP_NOT_REPRESENTABLE
	StatusTempFileError          Status = C.CAIRO_STATUS_TEMP_FILE_ERROR
	StatusInvalidStride          Status = C.CAIRO_STATUS_INVALID_STRIDE
	StatusFontTypeMismatch       Status = C.CAIRO_STATUS_FONT_TYPE_MISMATCH
	StatusUserFontImmutable      Status = C.CAIRO_STATUS_USER_FONT_IMMUTABLE
	StatusUserFontError          Status = C.CAIRO_STATUS_USER_FONT_ERROR
	StatusNegativeCount          Status = C.CAIRO_STATUS_NEGATIVE_COUNT
	StatusInvalidClusters        Status = C.CAIRO_STATUS_INVALID_CLUSTERS
	StatusInvalidSlant           Status = C.CAIRO_STATUS_INVALID_SLANT
	StatusInvalidWeight          Status = C.CAIRO_STATUS_INVALID_WEIGHT
	StatusInvalidSize            Status = C.CAIRO_STATUS_INVALID_SIZE
	StatusUserFontNotImplemented Status = C.CAIRO_STATUS_USER_FONT_NOT_IMPLEMENTED
	StatusDeviceTypeMismatch     Status = C.CAIRO_STATUS_DEVICE_TYPE_MISMATCH
	StatusDeviceError            Status = C.CAIRO_STATUS_DEVICE_ERROR
	StatusLastStatus             Status = C.CAIRO_STATUS_LAST_STATUS
)

func (this Status) c() C.cairo_status_t {
	return C.cairo_status_t(this)
}

func (this Status) String() string {
	cstr := C.cairo_status_to_string(this.c())
	return C.GoString(cstr)
}

//----------------------------------------------------------------------------
// Drawing Context
//----------------------------------------------------------------------------

// typedef             cairo_t;
type Context struct {
	C *C.cairo_t
}

func context_finalizer(this *Context) {
	C.cairo_set_user_data(this.C, &go_repr_cookie, nil, nil)
	C.cairo_destroy(this.C)
}

func ContextWrap(c *C.cairo_t, grab bool) *Context {
	go_repr := C.cairo_get_user_data(c, &go_repr_cookie)
	if go_repr != nil {
		return (*Context)(go_repr)
	}

	context := &Context{c}
	if grab {
		C.cairo_reference(c)
	}
	runtime.SetFinalizer(context, context_finalizer)

	status := C.cairo_set_user_data(c, &go_repr_cookie, unsafe.Pointer(context), nil)
	if status != C.CAIRO_STATUS_SUCCESS {
		panic("failed to set user data, out of memory?")
	}
	return context
}

// cairo_t *           cairo_create                        (cairo_surface_t *target);
func NewContext(target *Surface) *Context {
	return ContextWrap(C.cairo_create(target.C), false)
}

// cairo_t *           cairo_reference                     (cairo_t *cr);
// void                cairo_destroy                       (cairo_t *cr);

// cairo_status_t      cairo_status                        (cairo_t *cr);
func (this *Context) Status() Status {
	return Status(C.cairo_status(this.C))
}

// void                cairo_save                          (cairo_t *cr);
func (this *Context) Save() {
	C.cairo_save(this.C)
}

// void                cairo_restore                       (cairo_t *cr);
func (this *Context) Restore() {
	C.cairo_restore(this.C)
}

// cairo_surface_t *   cairo_get_target                    (cairo_t *cr);
func (this *Context) GetTarget() *Surface {
	return SurfaceWrap(C.cairo_get_target(this.C), true)
}

// void                cairo_push_group                    (cairo_t *cr);
func (this *Context) PushGroup() {
	C.cairo_push_group(this.C)
}

// void                cairo_push_group_with_content       (cairo_t *cr,
//                                                          cairo_content_t content);
func (this *Context) PushGroupWithContent(content Content) {
	C.cairo_push_group_with_content(this.C, content.c())
}

// cairo_pattern_t *   cairo_pop_group                     (cairo_t *cr);
func (this *Context) PopGroup() *Pattern {
	return (*Pattern)(PatternWrap(C.cairo_pop_group(this.C), false))
}

// void                cairo_pop_group_to_source           (cairo_t *cr);
func (this *Context) PopGroupToSource() {
	C.cairo_pop_group_to_source(this.C)
}

// cairo_surface_t *   cairo_get_group_target              (cairo_t *cr);
func (this *Context) GetGroupTarget() *Surface {
	return SurfaceWrap(C.cairo_get_group_target(this.C), true)
}

// void                cairo_set_source_rgb                (cairo_t *cr,
//                                                          double red,
//                                                          double green,
//                                                          double blue);
func (this *Context) SetSourceRGB(r, g, b float64) {
	C.cairo_set_source_rgb(this.C, C.double(r), C.double(g), C.double(b))
}

// void                cairo_set_source_rgba               (cairo_t *cr,
//                                                          double red,
//                                                          double green,
//                                                          double blue,
//                                                          double alpha);
func (this *Context) SetSourceRGBA(r, g, b, a float64) {
	C.cairo_set_source_rgba(this.C, C.double(r), C.double(g), C.double(b), C.double(a))
}

// void                cairo_set_source                    (cairo_t *cr,
//                                                          cairo_pattern_t *source);
func (this *Context) SetSource(source PatternLike) {
	C.cairo_set_source(this.C, source.InheritedFromCairoPattern())
}

// void                cairo_set_source_surface            (cairo_t *cr,
//                                                          cairo_surface_t *surface,
//                                                          double x,
//                                                          double y);
func (this *Context) SetSourceSurface(surface *Surface, x, y float64) {
	C.cairo_set_source_surface(this.C, surface.C, C.double(x), C.double(y))
}

// cairo_pattern_t *   cairo_get_source                    (cairo_t *cr);
func (this *Context) GetSource() *Pattern {
	return (*Pattern)(PatternWrap(C.cairo_get_source(this.C), true))
}

// enum                cairo_antialias_t;
type Antialias int

const (
	AntialiasDefault  Antialias = C.CAIRO_ANTIALIAS_DEFAULT
	AntialiasNone     Antialias = C.CAIRO_ANTIALIAS_NONE
	AntialiasGray     Antialias = C.CAIRO_ANTIALIAS_GRAY
	AntialiasSubpixel Antialias = C.CAIRO_ANTIALIAS_SUBPIXEL
)

func (this Antialias) c() C.cairo_antialias_t {
	return C.cairo_antialias_t(this)
}

// void                cairo_set_antialias                 (cairo_t *cr,
//                                                          cairo_antialias_t antialias);
func (this *Context) SetAntialias(antialias Antialias) {
	C.cairo_set_antialias(this.C, antialias.c())
}

// cairo_antialias_t   cairo_get_antialias                 (cairo_t *cr);
func (this *Context) GetAntialias() Antialias {
	return Antialias(C.cairo_get_antialias(this.C))
}

// void                cairo_set_dash                      (cairo_t *cr,
//                                                          const double *dashes,
//                                                          int num_dashes,
//                                                          double offset);
func (this *Context) SetDash(dashes []float64, offset float64) {
	var first *C.double
	if len(dashes) != 0 {
		first = (*C.double)(unsafe.Pointer(&dashes[0]))
	}
	C.cairo_set_dash(this.C, first, C.int(len(dashes)), C.double(offset))
}

// int                 cairo_get_dash_count                (cairo_t *cr);
func (this *Context) GetDashCount() int {
	return int(C.cairo_get_dash_count(this.C))
}

// void                cairo_get_dash                      (cairo_t *cr,
//                                                          double *dashes,
//                                                          double *offset);
func (this *Context) GetDash() ([]float64, float64) {
	var first *C.double
	var offset C.double
	var dashes []float64

	count := this.GetDashCount()
	if count != 0 {
		dashes = make([]float64, count)
		first = (*C.double)(unsafe.Pointer(&dashes[0]))
	}
	C.cairo_get_dash(this.C, first, &offset)
	return dashes, float64(offset)
}

// enum                cairo_fill_rule_t;
type FillRule int

const (
	FillRuleWinding FillRule = C.CAIRO_FILL_RULE_WINDING
	FillRuleEvenOdd FillRule = C.CAIRO_FILL_RULE_EVEN_ODD
)

func (this FillRule) c() C.cairo_fill_rule_t {
	return C.cairo_fill_rule_t(this)
}

// void                cairo_set_fill_rule                 (cairo_t *cr,
//                                                          cairo_fill_rule_t fill_rule);
func (this *Context) SetFillRule(fill_rule FillRule) {
	C.cairo_set_fill_rule(this.C, fill_rule.c())
}

// cairo_fill_rule_t   cairo_get_fill_rule                 (cairo_t *cr);
func (this *Context) GetFillRule() FillRule {
	return FillRule(C.cairo_get_fill_rule(this.C))
}

// enum                cairo_line_cap_t;
type LineCap int

const (
	LineCapButt   LineCap = C.CAIRO_LINE_CAP_BUTT
	LineCapRound  LineCap = C.CAIRO_LINE_CAP_ROUND
	LineCapSquare LineCap = C.CAIRO_LINE_CAP_SQUARE
)

func (this LineCap) c() C.cairo_line_cap_t {
	return C.cairo_line_cap_t(this)
}

// void                cairo_set_line_cap                  (cairo_t *cr,
//                                                          cairo_line_cap_t line_cap);
func (this *Context) SetLineCap(line_cap LineCap) {
	C.cairo_set_line_cap(this.C, line_cap.c())
}

// cairo_line_cap_t    cairo_get_line_cap                  (cairo_t *cr);
func (this *Context) GetLineCap() LineCap {
	return LineCap(C.cairo_get_line_cap(this.C))
}

// enum                cairo_line_join_t;
type LineJoin int

const (
	LineJoinMiter LineJoin = C.CAIRO_LINE_JOIN_MITER
	LineJoinRound LineJoin = C.CAIRO_LINE_JOIN_ROUND
	LineJoinBevel LineJoin = C.CAIRO_LINE_JOIN_BEVEL
)

func (this LineJoin) c() C.cairo_line_join_t {
	return C.cairo_line_join_t(this)
}

// void                cairo_set_line_join                 (cairo_t *cr,
//                                                          cairo_line_join_t line_join);
func (this *Context) SetLineJoin(line_join LineJoin) {
	C.cairo_set_line_join(this.C, line_join.c())
}

// cairo_line_join_t   cairo_get_line_join                 (cairo_t *cr);
func (this *Context) GetLineJoin() LineJoin {
	return LineJoin(C.cairo_get_line_join(this.C))
}

// void                cairo_set_line_width                (cairo_t *cr,
//                                                          double width);
func (this *Context) SetLineWidth(width float64) {
	C.cairo_set_line_width(this.C, C.double(width))
}

// double              cairo_get_line_width                (cairo_t *cr);
func (this *Context) GetLineWidth() float64 {
	return float64(C.cairo_get_line_width(this.C))
}

// void                cairo_set_miter_limit               (cairo_t *cr,
//                                                          double limit);
func (this *Context) SetMiterLimit(limit float64) {
	C.cairo_set_miter_limit(this.C, C.double(limit))
}

// double              cairo_get_miter_limit               (cairo_t *cr);
func (this *Context) GetMiterLimit() float64 {
	return float64(C.cairo_get_miter_limit(this.C))
}

// enum                cairo_operator_t;
type Operator int

const (
	OperatorClear         Operator = C.CAIRO_OPERATOR_CLEAR
	OperatorSource        Operator = C.CAIRO_OPERATOR_SOURCE
	OperatorOver          Operator = C.CAIRO_OPERATOR_OVER
	OperatorIn            Operator = C.CAIRO_OPERATOR_IN
	OperatorOut           Operator = C.CAIRO_OPERATOR_OUT
	OperatorAtop          Operator = C.CAIRO_OPERATOR_ATOP
	OperatorDest          Operator = C.CAIRO_OPERATOR_DEST
	OperatorDestOver      Operator = C.CAIRO_OPERATOR_DEST_OVER
	OperatorDestIn        Operator = C.CAIRO_OPERATOR_DEST_IN
	OperatorDestOut       Operator = C.CAIRO_OPERATOR_DEST_OUT
	OperatorDestAtop      Operator = C.CAIRO_OPERATOR_DEST_ATOP
	OperatorXor           Operator = C.CAIRO_OPERATOR_XOR
	OperatorAdd           Operator = C.CAIRO_OPERATOR_ADD
	OperatorSaturate      Operator = C.CAIRO_OPERATOR_SATURATE
	OperatorMultiply      Operator = C.CAIRO_OPERATOR_MULTIPLY
	OperatorScreen        Operator = C.CAIRO_OPERATOR_SCREEN
	OperatorOverlay       Operator = C.CAIRO_OPERATOR_OVERLAY
	OperatorDarken        Operator = C.CAIRO_OPERATOR_DARKEN
	OperatorLighten       Operator = C.CAIRO_OPERATOR_LIGHTEN
	OperatorColorDodge    Operator = C.CAIRO_OPERATOR_COLOR_DODGE
	OperatorColorBurn     Operator = C.CAIRO_OPERATOR_COLOR_BURN
	OperatorHardLight     Operator = C.CAIRO_OPERATOR_HARD_LIGHT
	OperatorSoftLight     Operator = C.CAIRO_OPERATOR_SOFT_LIGHT
	OperatorDifference    Operator = C.CAIRO_OPERATOR_DIFFERENCE
	OperatorExclusion     Operator = C.CAIRO_OPERATOR_EXCLUSION
	OperatorHslHue        Operator = C.CAIRO_OPERATOR_HSL_HUE
	OperatorHslSaturation Operator = C.CAIRO_OPERATOR_HSL_SATURATION
	OperatorHslColor      Operator = C.CAIRO_OPERATOR_HSL_COLOR
	OperatorHslLuminosity Operator = C.CAIRO_OPERATOR_HSL_LUMINOSITY
)

func (this Operator) c() C.cairo_operator_t {
	return C.cairo_operator_t(this)
}

// void                cairo_set_operator                  (cairo_t *cr,
//                                                          cairo_operator_t op);
func (this *Context) SetOperator(op Operator) {
	C.cairo_set_operator(this.C, op.c())
}

// cairo_operator_t    cairo_get_operator                  (cairo_t *cr);
func (this *Context) GetOperator() Operator {
	return Operator(C.cairo_get_operator(this.C))
}

// void                cairo_set_tolerance                 (cairo_t *cr,
//                                                          double tolerance);
func (this *Context) SetTolerance(tolerance float64) {
	C.cairo_set_tolerance(this.C, C.double(tolerance))
}

// double              cairo_get_tolerance                 (cairo_t *cr);
func (this *Context) GetTolerance() float64 {
	return float64(C.cairo_get_tolerance(this.C))
}

// void                cairo_clip                          (cairo_t *cr);
func (this *Context) Clip() {
	C.cairo_clip(this.C)
}

// void                cairo_clip_preserve                 (cairo_t *cr);
func (this *Context) ClipPreserve() {
	C.cairo_clip_preserve(this.C)
}

// void                cairo_clip_extents                  (cairo_t *cr,
//                                                          double *x1,
//                                                          double *y1,
//                                                          double *x2,
//                                                          double *y2);
func (this *Context) ClipExtents() (x1, y1, x2, y2 float64) {
	C.cairo_clip_extents(this.C,
		(*C.double)(unsafe.Pointer(&x1)),
		(*C.double)(unsafe.Pointer(&y1)),
		(*C.double)(unsafe.Pointer(&x2)),
		(*C.double)(unsafe.Pointer(&y2)))
	return
}

// cairo_bool_t        cairo_in_clip                       (cairo_t *cr,
//                                                          double x,
//                                                          double y);
func (this *Context) InClip(x, y float64) bool {
	return C.cairo_in_clip(this.C, C.double(x), C.double(y)) != 0
}

// void                cairo_reset_clip                    (cairo_t *cr);
func (this *Context) ResetClip() {
	C.cairo_reset_clip(this.C)
}

//                     cairo_rectangle_t;
type Rectangle struct {
	X, Y, Width, Height float64
}

//                     cairo_rectangle_list_t;
// void                cairo_rectangle_list_destroy        (cairo_rectangle_list_t *rectangle_list);

// cairo_rectangle_list_t * cairo_copy_clip_rectangle_list (cairo_t *cr);
func (this *Context) CopyClipRectangleList() ([]Rectangle, Status) {
	var slice []Rectangle
	var status Status

	rl := C.cairo_copy_clip_rectangle_list(this.C)
	if rl.num_rectangles > 0 {
		var slice_header reflect.SliceHeader
		slice_header.Data = uintptr(unsafe.Pointer(rl.rectangles))
		slice_header.Len = int(rl.num_rectangles)
		slice_header.Cap = int(rl.num_rectangles)
		slice_src := *(*[]Rectangle)(unsafe.Pointer(&slice_header))
		slice = make([]Rectangle, rl.num_rectangles)
		copy(slice, slice_src)
	}

	status = Status(rl.status)

	C.cairo_rectangle_list_destroy(rl)
	return slice, status
}

// void                cairo_fill                          (cairo_t *cr);
func (this *Context) Fill() {
	C.cairo_fill(this.C)
}

// void                cairo_fill_preserve                 (cairo_t *cr);
func (this *Context) FillPreserve() {
	C.cairo_fill_preserve(this.C)
}

// void                cairo_fill_extents                  (cairo_t *cr,
//                                                          double *x1,
//                                                          double *y1,
//                                                          double *x2,
//                                                          double *y2);
func (this *Context) FillExtents() (x1, y1, x2, y2 float64) {
	C.cairo_fill_extents(this.C,
		(*C.double)(unsafe.Pointer(&x1)),
		(*C.double)(unsafe.Pointer(&y1)),
		(*C.double)(unsafe.Pointer(&x2)),
		(*C.double)(unsafe.Pointer(&y2)))
	return
}

// cairo_bool_t        cairo_in_fill                       (cairo_t *cr,
//                                                          double x,
//                                                          double y);
func (this *Context) InFill(x, y float64) bool {
	return C.cairo_in_fill(this.C, C.double(x), C.double(y)) != 0
}

// void                cairo_mask                          (cairo_t *cr,
//                                                          cairo_pattern_t *pattern);
func (this *Context) Mask(pattern PatternLike) {
	C.cairo_mask(this.C, pattern.InheritedFromCairoPattern())
}

// void                cairo_mask_surface                  (cairo_t *cr,
//                                                          cairo_surface_t *surface,
//                                                          double surface_x,
//                                                          double surface_y);
func (this *Context) MaskSurface(surface *Surface, surface_x, surface_y float64) {
	C.cairo_mask_surface(this.C, surface.C, C.double(surface_x), C.double(surface_y))
}

// void                cairo_paint                         (cairo_t *cr);
func (this *Context) Paint() {
	C.cairo_paint(this.C)
}

// void                cairo_paint_with_alpha              (cairo_t *cr,
//                                                          double alpha);
func (this *Context) PaintWithAlpha(alpha float64) {
	C.cairo_paint_with_alpha(this.C, C.double(alpha))
}

// void                cairo_stroke                        (cairo_t *cr);
func (this *Context) Stroke() {
	C.cairo_stroke(this.C)
}

// void                cairo_stroke_preserve               (cairo_t *cr);
func (this *Context) StrokePreserve() {
	C.cairo_stroke_preserve(this.C)
}

// void                cairo_stroke_extents                (cairo_t *cr,
//                                                          double *x1,
//                                                          double *y1,
//                                                          double *x2,
//                                                          double *y2);
func (this *Context) StrokeExtents() (x1, y1, x2, y2 float64) {
	C.cairo_stroke_extents(this.C,
		(*C.double)(unsafe.Pointer(&x1)),
		(*C.double)(unsafe.Pointer(&y1)),
		(*C.double)(unsafe.Pointer(&x2)),
		(*C.double)(unsafe.Pointer(&y2)))
	return
}

// cairo_bool_t        cairo_in_stroke                     (cairo_t *cr,
//                                                          double x,
//                                                          double y);
func (this *Context) InStroke(x, y float64) bool {
	return C.cairo_in_stroke(this.C, C.double(x), C.double(y)) != 0
}

// void                cairo_copy_page                     (cairo_t *cr);
func (this *Context) CopyPage() {
	C.cairo_copy_page(this.C)
}

// void                cairo_show_page                     (cairo_t *cr);
func (this *Context) ShowPage() {
	C.cairo_show_page(this.C)
}

// unsigned int        cairo_get_reference_count           (cairo_t *cr);
// cairo_status_t      cairo_set_user_data                 (cairo_t *cr,
//                                                          const cairo_user_data_key_t *key,
//                                                          void *user_data,
//                                                          cairo_destroy_func_t destroy);
// void *              cairo_get_user_data                 (cairo_t *cr,
//                                                          const cairo_user_data_key_t *key);

//----------------------------------------------------------------------------
// Paths
//----------------------------------------------------------------------------

//                     cairo_path_t;
type Path struct {
	C *C.cairo_path_t
}

func path_finalizer(this *Path) {
	C.cairo_path_destroy(this.C)
}

func PathWrap(c *C.cairo_path_t) *Path {
	path := &Path{c}
	runtime.SetFinalizer(path, path_finalizer)
	return path
}

// TODO: Implement?
// union               cairo_path_data_t;

// enum                cairo_path_data_type_t;
type PathDataType int

const (
	PathMoveTo    PathDataType = C.CAIRO_PATH_MOVE_TO
	PathLineTo    PathDataType = C.CAIRO_PATH_LINE_TO
	PathCurveTo   PathDataType = C.CAIRO_PATH_CURVE_TO
	PathClosePath PathDataType = C.CAIRO_PATH_CLOSE_PATH
)

// cairo_path_t *      cairo_copy_path                     (cairo_t *cr);
func (this *Context) CopyPath() *Path {
	return PathWrap(C.cairo_copy_path(this.C))
}

// cairo_path_t *      cairo_copy_path_flat                (cairo_t *cr);
func (this *Context) CopyPathFlat() *Path {
	return PathWrap(C.cairo_copy_path_flat(this.C))
}

// void                cairo_path_destroy                  (cairo_path_t *path);

// void                cairo_append_path                   (cairo_t *cr,
//                                                          const cairo_path_t *path);
func (this *Context) AppendPath(path *Path) {
	C.cairo_append_path(this.C, path.C)
}

// cairo_bool_t        cairo_has_current_point             (cairo_t *cr);
func (this *Context) HasCurrentPoint() bool {
	return C.cairo_has_current_point(this.C) != 0
}

// void                cairo_get_current_point             (cairo_t *cr,
//                                                          double *x,
//                                                          double *y);
func (this *Context) GetCurrentPoint() (x, y float64) {
	C.cairo_get_current_point(this.C,
		(*C.double)(unsafe.Pointer(&x)),
		(*C.double)(unsafe.Pointer(&y)))
	return
}

// void                cairo_new_path                      (cairo_t *cr);
func (this *Context) NewPath() {
	C.cairo_new_path(this.C)
}

// void                cairo_new_sub_path                  (cairo_t *cr);
func (this *Context) NewSubPath() {
	C.cairo_new_sub_path(this.C)
}

// void                cairo_close_path                    (cairo_t *cr);
func (this *Context) ClosePath() {
	C.cairo_close_path(this.C)
}

// void                cairo_arc                           (cairo_t *cr,
//                                                          double xc,
//                                                          double yc,
//                                                          double radius,
//                                                          double angle1,
//                                                          double angle2);
func (this *Context) Arc(xc, yc, radius, angle1, angle2 float64) {
	C.cairo_arc(this.C, C.double(xc), C.double(yc), C.double(radius), C.double(angle1), C.double(angle2))
}

// void                cairo_arc_negative                  (cairo_t *cr,
//                                                          double xc,
//                                                          double yc,
//                                                          double radius,
//                                                          double angle1,
//                                                          double angle2);
func (this *Context) ArcNegative(xc, yc, radius, angle1, angle2 float64) {
	C.cairo_arc_negative(this.C, C.double(xc), C.double(yc), C.double(radius), C.double(angle1), C.double(angle2))
}

// void                cairo_curve_to                      (cairo_t *cr,
//                                                          double x1,
//                                                          double y1,
//                                                          double x2,
//                                                          double y2,
//                                                          double x3,
//                                                          double y3);
func (this *Context) CurveTo(x1, y1, x2, y2, x3, y3 float64) {
	C.cairo_curve_to(this.C, C.double(x1), C.double(y1), C.double(x2), C.double(y2), C.double(x3), C.double(y3))
}

// void                cairo_line_to                       (cairo_t *cr,
//                                                          double x,
//                                                          double y);
func (this *Context) LineTo(x, y float64) {
	C.cairo_line_to(this.C, C.double(x), C.double(y))
}

// void                cairo_move_to                       (cairo_t *cr,
//                                                          double x,
//                                                          double y);
func (this *Context) MoveTo(x, y float64) {
	C.cairo_move_to(this.C, C.double(x), C.double(y))
}

// void                cairo_rectangle                     (cairo_t *cr,
//                                                          double x,
//                                                          double y,
//                                                          double width,
//                                                          double height);
func (this *Context) Rectangle(x, y, width, height float64) {
	C.cairo_rectangle(this.C, C.double(x), C.double(y), C.double(width), C.double(height))
}

// TODO: Implement
// void                cairo_glyph_path                    (cairo_t *cr,
//                                                          const cairo_glyph_t *glyphs,
//                                                          int num_glyphs);

// void                cairo_text_path                     (cairo_t *cr,
//                                                          const char *utf8);
func (this *Context) TextPath(utf8 string) {
	utf8c := C.CString(utf8)
	C.cairo_text_path(this.C, utf8c)
	C.free(unsafe.Pointer(utf8c))
}

// void                cairo_rel_curve_to                  (cairo_t *cr,
//                                                          double dx1,
//                                                          double dy1,
//                                                          double dx2,
//                                                          double dy2,
//                                                          double dx3,
//                                                          double dy3);
func (this *Context) RelCurveTo(dx1, dy1, dx2, dy2, dx3, dy3 float64) {
	C.cairo_rel_curve_to(this.C, C.double(dx1), C.double(dy1), C.double(dx2), C.double(dy2), C.double(dx3), C.double(dy3))
}

// void                cairo_rel_line_to                   (cairo_t *cr,
//                                                          double dx,
//                                                          double dy);
func (this *Context) RelLineTo(dx, dy float64) {
	C.cairo_rel_line_to(this.C, C.double(dx), C.double(dy))
}

// void                cairo_rel_move_to                   (cairo_t *cr,
//                                                          double dx,
//                                                          double dy);
func (this *Context) RelMoveTo(dx, dy float64) {
	C.cairo_rel_move_to(this.C, C.double(dx), C.double(dy))
}

// void                cairo_path_extents                  (cairo_t *cr,
//                                                          double *x1,
//                                                          double *y1,
//                                                          double *x2,
//                                                          double *y2);
func (this *Context) PathExtents() (x1, y1, x2, y2 float64) {
	C.cairo_path_extents(this.C,
		(*C.double)(unsafe.Pointer(&x1)),
		(*C.double)(unsafe.Pointer(&y1)),
		(*C.double)(unsafe.Pointer(&x2)),
		(*C.double)(unsafe.Pointer(&y2)))
	return
}

//----------------------------------------------------------------------------
// Pattern
// 	SolidPattern
// 	SurfacePattern
// 	Gradient
// 		LinearGradient
// 		RadialGradient
//----------------------------------------------------------------------------

type PatternLike interface {
	InheritedFromCairoPattern() *C.cairo_pattern_t
}

// typedef             cairo_pattern_t;
type Pattern struct{ C *C.cairo_pattern_t }
type SolidPattern struct{ Pattern }
type SurfacePattern struct{ Pattern }
type Gradient struct{ Pattern }
type LinearGradient struct{ Gradient }
type RadialGradient struct{ Gradient }

func pattern_finalizer(this *Pattern) {
	C.cairo_pattern_set_user_data(this.C, &go_repr_cookie, nil, nil)
	C.cairo_pattern_destroy(this.C)
}

func PatternWrap(c *C.cairo_pattern_t, grab bool) unsafe.Pointer {
	go_repr := C.cairo_pattern_get_user_data(c, &go_repr_cookie)
	if go_repr != nil {
		return unsafe.Pointer(go_repr)
	}

	pattern := &Pattern{c}
	if grab {
		C.cairo_pattern_reference(c)
	}
	runtime.SetFinalizer(pattern, pattern_finalizer)

	status := C.cairo_pattern_set_user_data(c, &go_repr_cookie, unsafe.Pointer(pattern), nil)
	if status != C.CAIRO_STATUS_SUCCESS {
		panic("failed to set user data, out of memory?")
	}
	return unsafe.Pointer(pattern)
}

func ensure_pattern_type(c *C.cairo_pattern_t, types ...PatternType) {
	for _, t := range types {
		if C.cairo_pattern_get_type(c) == t.c() {
			return
		}
	}
	panic("unexpected pattern type")
}

func ToPattern(like PatternLike) *Pattern {
	return (*Pattern)(PatternWrap(like.InheritedFromCairoPattern(), true))
}

func ToSolidPattern(like PatternLike) *SolidPattern {
	c := like.InheritedFromCairoPattern()
	ensure_pattern_type(c, PatternTypeSolid)
	return (*SolidPattern)(PatternWrap(c, true))
}

func ToSurfacePattern(like PatternLike) *SurfacePattern {
	c := like.InheritedFromCairoPattern()
	ensure_pattern_type(c, PatternTypeSurface)
	return (*SurfacePattern)(PatternWrap(c, true))
}

func ToGradient(like PatternLike) *Gradient {
	c := like.InheritedFromCairoPattern()
	ensure_pattern_type(c, PatternTypeLinear, PatternTypeRadial)
	return (*Gradient)(PatternWrap(c, true))
}

func ToLinearGradient(like PatternLike) *LinearGradient {
	c := like.InheritedFromCairoPattern()
	ensure_pattern_type(c, PatternTypeLinear)
	return (*LinearGradient)(PatternWrap(c, true))
}

func ToRadialGradient(like PatternLike) *RadialGradient {
	c := like.InheritedFromCairoPattern()
	ensure_pattern_type(c, PatternTypeRadial)
	return (*RadialGradient)(PatternWrap(c, true))
}

func (this *Pattern) InheritedFromCairoPattern() *C.cairo_pattern_t {
	return this.C
}

// void                cairo_pattern_add_color_stop_rgb    (cairo_pattern_t *pattern,
//                                                          double offset,
//                                                          double red,
//                                                          double green,
//                                                          double blue);
func (this *Gradient) AddColorStopRGB(offset, red, green, blue float64) {
	C.cairo_pattern_add_color_stop_rgb(this.C,
		C.double(offset), C.double(red), C.double(green), C.double(blue))
}

// void                cairo_pattern_add_color_stop_rgba   (cairo_pattern_t *pattern,
//                                                          double offset,
//                                                          double red,
//                                                          double green,
//                                                          double blue,
//                                                          double alpha);
func (this *Gradient) AddColorStopRGBA(offset, red, green, blue, alpha float64) {
	C.cairo_pattern_add_color_stop_rgba(this.C,
		C.double(offset), C.double(red), C.double(green), C.double(blue), C.double(alpha))
}

// cairo_status_t      cairo_pattern_get_color_stop_count  (cairo_pattern_t *pattern,
//                                                          int *count);
func (this *Gradient) GetColorStopCount() (count int) {
	C.cairo_pattern_get_color_stop_count(this.C,
		(*C.int)(unsafe.Pointer(&count)))
	return
}

// cairo_status_t      cairo_pattern_get_color_stop_rgba   (cairo_pattern_t *pattern,
//                                                          int index,
//                                                          double *offset,
//                                                          double *red,
//                                                          double *green,
//                                                          double *blue,
//                                                          double *alpha);
func (this *Gradient) GetColorStopRGBA(index int) (offset, red, green, blue, alpha float64) {
	C.cairo_pattern_get_color_stop_rgba(this.C, C.int(index),
		(*C.double)(unsafe.Pointer(&offset)),
		(*C.double)(unsafe.Pointer(&red)),
		(*C.double)(unsafe.Pointer(&green)),
		(*C.double)(unsafe.Pointer(&blue)),
		(*C.double)(unsafe.Pointer(&alpha)))
	return
}

// cairo_pattern_t *   cairo_pattern_create_rgb            (double red,
//                                                          double green,
//                                                          double blue);
func NewSolidPatternRGB(red, green, blue float64) *SolidPattern {
	return (*SolidPattern)(PatternWrap(C.cairo_pattern_create_rgb(C.double(red), C.double(green), C.double(blue)), false))
}

// cairo_pattern_t *   cairo_pattern_create_rgba           (double red,
//                                                          double green,
//                                                          double blue,
//                                                          double alpha);
func NewSolidPatternRGBA(red, green, blue, alpha float64) *SolidPattern {
	return (*SolidPattern)(PatternWrap(C.cairo_pattern_create_rgba(C.double(red), C.double(green), C.double(blue), C.double(alpha)), false))
}

// cairo_status_t      cairo_pattern_get_rgba              (cairo_pattern_t *pattern,
//                                                          double *red,
//                                                          double *green,
//                                                          double *blue,
//                                                          double *alpha);
func (this *SolidPattern) GetRGBA() (red, green, blue, alpha float64) {
	C.cairo_pattern_get_rgba(this.C,
		(*C.double)(unsafe.Pointer(&red)),
		(*C.double)(unsafe.Pointer(&green)),
		(*C.double)(unsafe.Pointer(&blue)),
		(*C.double)(unsafe.Pointer(&alpha)))
	return
}

// cairo_pattern_t *   cairo_pattern_create_for_surface    (cairo_surface_t *surface);
func NewSurfacePattern(surface *Surface) *SurfacePattern {
	return (*SurfacePattern)(PatternWrap(C.cairo_pattern_create_for_surface(surface.C), false))
}

// cairo_status_t      cairo_pattern_get_surface           (cairo_pattern_t *pattern,
//                                                          cairo_surface_t **surface);
func (this *SurfacePattern) GetSurface() *Surface {
	var surfacec *C.cairo_surface_t
	C.cairo_pattern_get_surface(this.C, &surfacec)
	surface := SurfaceWrap(surfacec, true)
	return surface
}

// cairo_pattern_t *   cairo_pattern_create_linear         (double x0,
//                                                          double y0,
//                                                          double x1,
//                                                          double y1);
func NewLinearGradient(x0, y0, x1, y1 float64) *LinearGradient {
	return (*LinearGradient)(PatternWrap(C.cairo_pattern_create_linear(
		C.double(x0), C.double(y0), C.double(x1), C.double(y1)), false))
}

// cairo_status_t      cairo_pattern_get_linear_points     (cairo_pattern_t *pattern,
//                                                          double *x0,
//                                                          double *y0,
//                                                          double *x1,
//                                                          double *y1);
func (this *LinearGradient) GetLinearPoints() (x0, y0, x1, y1 float64) {
	C.cairo_pattern_get_linear_points(this.C,
		(*C.double)(unsafe.Pointer(&x0)),
		(*C.double)(unsafe.Pointer(&y0)),
		(*C.double)(unsafe.Pointer(&x1)),
		(*C.double)(unsafe.Pointer(&y1)))
	return
}

// cairo_pattern_t *   cairo_pattern_create_radial         (double cx0,
//                                                          double cy0,
//                                                          double radius0,
//                                                          double cx1,
//                                                          double cy1,
//                                                          double radius1);
func NewRadialGradient(cx0, cy0, radius0, cx1, cy1, radius1 float64) *RadialGradient {
	return (*RadialGradient)(PatternWrap(C.cairo_pattern_create_radial(
		C.double(cx0), C.double(cy0), C.double(radius0), C.double(cx1), C.double(cy1), C.double(radius1)), false))
}

// cairo_status_t      cairo_pattern_get_radial_circles    (cairo_pattern_t *pattern,
//                                                          double *x0,
//                                                          double *y0,
//                                                          double *r0,
//                                                          double *x1,
//                                                          double *y1,
//                                                          double *r1);
func (this *RadialGradient) GetRadialCircles() (x0, y0, r0, x1, y1, r1 float64) {
	C.cairo_pattern_get_radial_circles(this.C,
		(*C.double)(unsafe.Pointer(&x0)),
		(*C.double)(unsafe.Pointer(&y0)),
		(*C.double)(unsafe.Pointer(&r0)),
		(*C.double)(unsafe.Pointer(&x1)),
		(*C.double)(unsafe.Pointer(&y1)),
		(*C.double)(unsafe.Pointer(&r1)))
	return
}

// cairo_pattern_t *   cairo_pattern_reference             (cairo_pattern_t *pattern);
// void                cairo_pattern_destroy               (cairo_pattern_t *pattern);

// cairo_status_t      cairo_pattern_status                (cairo_pattern_t *pattern);
func (this *Pattern) Status() Status {
	return Status(C.cairo_pattern_status(this.C))
}

// enum                cairo_extend_t;
type Extend int

const (
	ExtendNone    Extend = C.CAIRO_EXTEND_NONE
	ExtendRepeat  Extend = C.CAIRO_EXTEND_REPEAT
	ExtendReflect Extend = C.CAIRO_EXTEND_REFLECT
	ExtendPad     Extend = C.CAIRO_EXTEND_PAD
)

func (this Extend) c() C.cairo_extend_t {
	return C.cairo_extend_t(this)
}

// void                cairo_pattern_set_extend            (cairo_pattern_t *pattern,
//                                                          cairo_extend_t extend);
func (this *SurfacePattern) SetExtend(extend Extend) {
	C.cairo_pattern_set_extend(this.C, extend.c())
}

// cairo_extend_t      cairo_pattern_get_extend            (cairo_pattern_t *pattern);
func (this *SurfacePattern) GetExtend() Extend {
	return Extend(C.cairo_pattern_get_extend(this.C))
}

// enum                cairo_filter_t;
type Filter int

const (
	FilterFast     Filter = C.CAIRO_FILTER_FAST
	FilterGood     Filter = C.CAIRO_FILTER_GOOD
	FilterBest     Filter = C.CAIRO_FILTER_BEST
	FilterNearest  Filter = C.CAIRO_FILTER_NEAREST
	FilterBilinear Filter = C.CAIRO_FILTER_BILINEAR
	FilterGaussian Filter = C.CAIRO_FILTER_GAUSSIAN
)

func (this Filter) c() C.cairo_filter_t {
	return C.cairo_filter_t(this)
}

// void                cairo_pattern_set_filter            (cairo_pattern_t *pattern,
//                                                          cairo_filter_t filter);
func (this *SurfacePattern) SetFilter(filter Filter) {
	C.cairo_pattern_set_filter(this.C, filter.c())
}

// cairo_filter_t      cairo_pattern_get_filter            (cairo_pattern_t *pattern);
func (this *SurfacePattern) GetFilter() Filter {
	return Filter(C.cairo_pattern_get_filter(this.C))
}

// void                cairo_pattern_set_matrix            (cairo_pattern_t *pattern,
//                                                          const cairo_matrix_t *matrix);
func (this *Pattern) SetMatrix(matrix *Matrix) {
	C.cairo_pattern_set_matrix(this.C, matrix.c())
}

// void                cairo_pattern_get_matrix            (cairo_pattern_t *pattern,
//                                                          cairo_matrix_t *matrix);
func (this *Pattern) GetMatrix() Matrix {
	var matrix C.cairo_matrix_t
	C.cairo_pattern_get_matrix(this.C, &matrix)
	return *(*Matrix)(unsafe.Pointer(&matrix))
}

// enum                cairo_pattern_type_t;
type PatternType int

const (
	PatternTypeSolid   PatternType = C.CAIRO_PATTERN_TYPE_SOLID
	PatternTypeSurface PatternType = C.CAIRO_PATTERN_TYPE_SURFACE
	PatternTypeLinear  PatternType = C.CAIRO_PATTERN_TYPE_LINEAR
	PatternTypeRadial  PatternType = C.CAIRO_PATTERN_TYPE_RADIAL
)

func (this PatternType) c() C.cairo_pattern_type_t {
	return C.cairo_pattern_type_t(this)
}

// cairo_pattern_type_t  cairo_pattern_get_type            (cairo_pattern_t *pattern);
func (this *Pattern) GetType() PatternType {
	return PatternType(C.cairo_pattern_get_type(this.C))
}

// TODO: Implement these
// unsigned int        cairo_pattern_get_reference_count   (cairo_pattern_t *pattern);
// cairo_status_t      cairo_pattern_set_user_data         (cairo_pattern_t *pattern,
//                                                          const cairo_user_data_key_t *key,
//                                                          void *user_data,
//                                                          cairo_destroy_func_t destroy);
// void *              cairo_pattern_get_user_data         (cairo_pattern_t *pattern,
//                                                          const cairo_user_data_key_t *key);

//----------------------------------------------------------------------------
// Regions
//----------------------------------------------------------------------------

// typedef             cairo_region_t;
type Region struct {
	C *C.cairo_region_t
}

func region_finalizer(this *Region) {
	C.cairo_region_destroy(this.C)
}

func RegionWrap(c *C.cairo_region_t, grab bool) *Region {
	if grab {
		C.cairo_region_reference(c)
	}
	region := &Region{c}
	runtime.SetFinalizer(region, region_finalizer)
	return region
}

// cairo_region_t *    cairo_region_create                 (void);
func NewRegion() *Region {
	return RegionWrap(C.cairo_region_create(), false)
}

// cairo_region_t *    cairo_region_create_rectangle       (const cairo_rectangle_int_t *rectangle);
func NewRegionRectangle(rectangle *RectangleInt) *Region {
	return RegionWrap(C.cairo_region_create_rectangle(rectangle.c()), false)
}

// cairo_region_t *    cairo_region_create_rectangles      (const cairo_rectangle_int_t *rects,
//                                                          int count);
func NewRegionRectangles(rects []RectangleInt) *Region {
	var first *C.cairo_rectangle_int_t
	count := C.int(len(rects))
	if count > 0 {
		first = rects[0].c()
	}
	return RegionWrap(C.cairo_region_create_rectangles(first, count), false)
}

// cairo_region_t *    cairo_region_copy                   (const cairo_region_t *original);
func (this *Region) Copy() *Region {
	return RegionWrap(C.cairo_region_copy(this.C), false)
}

// cairo_region_t *    cairo_region_reference              (cairo_region_t *region);
// void                cairo_region_destroy                (cairo_region_t *region);

// cairo_status_t      cairo_region_status                 (const cairo_region_t *region);
func (this *Region) Status() Status {
	return Status(C.cairo_region_status(this.C))
}

// void                cairo_region_get_extents            (const cairo_region_t *region,
//                                                          cairo_rectangle_int_t *extents);
func (this *Region) GetExtents() (extents RectangleInt) {
	C.cairo_region_get_extents(this.C, extents.c())
	return
}

// int                 cairo_region_num_rectangles         (const cairo_region_t *region);
func (this *Region) NumRectangles() int {
	return int(C.cairo_region_num_rectangles(this.C))
}

// void                cairo_region_get_rectangle          (const cairo_region_t *region,
//                                                          int nth,
//                                                          cairo_rectangle_int_t *rectangle);
func (this *Region) GetRectangle(nth int) (rectangle RectangleInt) {
	C.cairo_region_get_rectangle(this.C, C.int(nth), rectangle.c())
	return
}

// cairo_bool_t        cairo_region_is_empty               (const cairo_region_t *region);
func (this *Region) IsEmpty() bool {
	return C.cairo_region_is_empty(this.C) != 0
}

// cairo_bool_t        cairo_region_contains_point         (const cairo_region_t *region,
//                                                          int x,
//                                                          int y);
func (this *Region) ContainsPoint(x, y int) bool {
	return C.cairo_region_contains_point(this.C, C.int(x), C.int(y)) != 0
}

// enum                cairo_region_overlap_t;
type RegionOverlap int

const (
	RegionOverlapIn   RegionOverlap = C.CAIRO_REGION_OVERLAP_IN   /* completely inside region */
	RegionOverlapOut  RegionOverlap = C.CAIRO_REGION_OVERLAP_OUT  /* completely outside region */
	RegionOverlapPart RegionOverlap = C.CAIRO_REGION_OVERLAP_PART /* partly inside region */
)

// cairo_region_overlap_t  cairo_region_contains_rectangle (const cairo_region_t *region,
//                                                          const cairo_rectangle_int_t *rectangle);
func (this *Region) ContainsRectangle(rectangle *RectangleInt) RegionOverlap {
	return RegionOverlap(C.cairo_region_contains_rectangle(this.C, rectangle.c()))
}

// cairo_bool_t        cairo_region_equal                  (const cairo_region_t *a,
//                                                          const cairo_region_t *b);
func (this *Region) Equal(b *Region) bool {
	return C.cairo_region_equal(this.C, b.C) != 0
}

// void                cairo_region_translate              (cairo_region_t *region,
//                                                          int dx,
//                                                          int dy);
func (this *Region) Translate(dx, dy int) {
	C.cairo_region_translate(this.C, C.int(dx), C.int(dy))
}

// cairo_status_t      cairo_region_intersect              (cairo_region_t *dst,
//                                                          const cairo_region_t *other);
func (this *Region) Intersect(other *Region) Status {
	return Status(C.cairo_region_intersect(this.C, other.C))
}

// cairo_status_t      cairo_region_intersect_rectangle    (cairo_region_t *dst,
//                                                          const cairo_rectangle_int_t *rectangle);
func (this *Region) IntersectRectangle(rectangle *RectangleInt) Status {
	return Status(C.cairo_region_intersect_rectangle(this.C, rectangle.c()))
}

// cairo_status_t      cairo_region_subtract               (cairo_region_t *dst,
//                                                          const cairo_region_t *other);
func (this *Region) Subtract(other *Region) Status {
	return Status(C.cairo_region_subtract(this.C, other.C))
}

// cairo_status_t      cairo_region_subtract_rectangle     (cairo_region_t *dst,
//                                                          const cairo_rectangle_int_t *rectangle);
func (this *Region) SubtractRectangle(rectangle *RectangleInt) Status {
	return Status(C.cairo_region_subtract_rectangle(this.C, rectangle.c()))
}

// cairo_status_t      cairo_region_union                  (cairo_region_t *dst,
//                                                          const cairo_region_t *other);
func (this *Region) Union(other *Region) Status {
	return Status(C.cairo_region_union(this.C, other.C))
}

// cairo_status_t      cairo_region_union_rectangle        (cairo_region_t *dst,
//                                                          const cairo_rectangle_int_t *rectangle);
func (this *Region) UnionRectangle(rectangle *RectangleInt) Status {
	return Status(C.cairo_region_union_rectangle(this.C, rectangle.c()))
}

// cairo_status_t      cairo_region_xor                    (cairo_region_t *dst,
//                                                          const cairo_region_t *other);
func (this *Region) Xor(other *Region) Status {
	return Status(C.cairo_region_xor(this.C, other.C))
}

// cairo_status_t      cairo_region_xor_rectangle          (cairo_region_t *dst,
//                                                          const cairo_rectangle_int_t *rectangle);
func (this *Region) XorRectangle(rectangle *RectangleInt) Status {
	return Status(C.cairo_region_xor_rectangle(this.C, rectangle.c()))
}

//----------------------------------------------------------------------------
// Transformations
//----------------------------------------------------------------------------

// void                cairo_translate                     (cairo_t *cr,
//                                                          double tx,
//                                                          double ty);
func (this *Context) Translate(tx, ty float64) {
	C.cairo_translate(this.C, C.double(tx), C.double(ty))
}

// void                cairo_scale                         (cairo_t *cr,
//                                                          double sx,
//                                                          double sy);
func (this *Context) Scale(sx, sy float64) {
	C.cairo_scale(this.C, C.double(sx), C.double(sy))
}

// void                cairo_rotate                        (cairo_t *cr,
//                                                          double angle);
func (this *Context) Rotate(angle float64) {
	C.cairo_rotate(this.C, C.double(angle))
}

// void                cairo_transform                     (cairo_t *cr,
//                                                          const cairo_matrix_t *matrix);
func (this *Context) Transform(matrix *Matrix) {
	C.cairo_transform(this.C, matrix.c())
}

// void                cairo_set_matrix                    (cairo_t *cr,
//                                                          const cairo_matrix_t *matrix);
func (this *Context) SetMatrix(matrix *Matrix) {
	C.cairo_set_matrix(this.C, matrix.c())
}

// void                cairo_get_matrix                    (cairo_t *cr,
//                                                          cairo_matrix_t *matrix);
func (this *Context) GetMatrix() Matrix {
	var matrix C.cairo_matrix_t
	C.cairo_get_matrix(this.C, &matrix)
	return *(*Matrix)(unsafe.Pointer(&matrix))
}

// void                cairo_identity_matrix               (cairo_t *cr);
func (this *Context) IdentityMatrix() {
	C.cairo_identity_matrix(this.C)
}

// void                cairo_user_to_device                (cairo_t *cr,
//                                                          double *x,
//                                                          double *y);
func (this *Context) UserToDevice(x, y float64) (float64, float64) {
	C.cairo_user_to_device(this.C,
		(*C.double)(unsafe.Pointer(&x)),
		(*C.double)(unsafe.Pointer(&y)))
	return x, y
}

// void                cairo_user_to_device_distance       (cairo_t *cr,
//                                                          double *dx,
//                                                          double *dy);
func (this *Context) UserToDeviceDistance(dx, dy float64) (float64, float64) {
	C.cairo_user_to_device_distance(this.C,
		(*C.double)(unsafe.Pointer(&dx)),
		(*C.double)(unsafe.Pointer(&dy)))
	return dx, dy
}

// void                cairo_device_to_user                (cairo_t *cr,
//                                                          double *x,
//                                                          double *y);
func (this *Context) DeviceToUser(x, y float64) (float64, float64) {
	C.cairo_device_to_user(this.C,
		(*C.double)(unsafe.Pointer(&x)),
		(*C.double)(unsafe.Pointer(&y)))
	return x, y
}

// void                cairo_device_to_user_distance       (cairo_t *cr,
//                                                          double *dx,
//                                                          double *dy);
func (this *Context) DeviceToUserDistance(dx, dy float64) (float64, float64) {
	C.cairo_device_to_user_distance(this.C,
		(*C.double)(unsafe.Pointer(&dx)),
		(*C.double)(unsafe.Pointer(&dy)))
	return dx, dy
}

//----------------------------------------------------------------------------
// Text
//----------------------------------------------------------------------------

//                     cairo_glyph_t;
// type Glyph (see types_$GOARCH.go)

func (this *Glyph) c() *C.cairo_glyph_t {
	return (*C.cairo_glyph_t)(unsafe.Pointer(this))
}

// enum                cairo_font_slant_t;
type FontSlant int

const (
	FontSlantNormal  FontSlant = C.CAIRO_FONT_SLANT_NORMAL
	FontSlantItalic  FontSlant = C.CAIRO_FONT_SLANT_ITALIC
	FontSlantOblique FontSlant = C.CAIRO_FONT_SLANT_OBLIQUE
)

func (this FontSlant) c() C.cairo_font_slant_t {
	return C.cairo_font_slant_t(this)
}

// enum                cairo_font_weight_t;
type FontWeight int

const (
	FontWeightNormal FontWeight = C.CAIRO_FONT_WEIGHT_NORMAL
	FontWeightBold   FontWeight = C.CAIRO_FONT_WEIGHT_BOLD
)

func (this FontWeight) c() C.cairo_font_weight_t {
	return C.cairo_font_weight_t(this)
}

//                     cairo_text_cluster_t;
type TextCluster struct {
	NumBytes  int32
	NumGlyphs int32
}

// enum                cairo_text_cluster_flags_t;
type TextClusterFlags int

const (
	TextClusterFlagBackward TextClusterFlags = C.CAIRO_TEXT_CLUSTER_FLAG_BACKWARD
)

// void                cairo_select_font_face              (cairo_t *cr,
//                                                          const char *family,
//                                                          cairo_font_slant_t slant,
//                                                          cairo_font_weight_t weight);
func (this *Context) SelectFontFace(family string, slant FontSlant, weight FontWeight) {
	cfamily := C.CString(family)
	C.cairo_select_font_face(this.C, cfamily, slant.c(), weight.c())
	C.free(unsafe.Pointer(cfamily))
}

// void                cairo_set_font_size                 (cairo_t *cr,
//                                                          double size);
func (this *Context) SetFontSize(size float64) {
	C.cairo_set_font_size(this.C, C.double(size))
}

// void                cairo_set_font_matrix               (cairo_t *cr,
//                                                          const cairo_matrix_t *matrix);
func (this *Context) SetFontMatrix(matrix *Matrix) {
	C.cairo_set_font_matrix(this.C, matrix.c())
}

// void                cairo_get_font_matrix               (cairo_t *cr,
//                                                          cairo_matrix_t *matrix);
func (this *Context) GetFontMatrix() Matrix {
	var matrix Matrix
	C.cairo_get_font_matrix(this.C, matrix.c())
	return matrix
}

// TODO: Implement these
// void                cairo_set_font_options              (cairo_t *cr,
//                                                          const cairo_font_options_t *options);
// void                cairo_get_font_options              (cairo_t *cr,
//                                                          cairo_font_options_t *options);
// void                cairo_set_font_face                 (cairo_t *cr,
//                                                          cairo_font_face_t *font_face);
// cairo_font_face_t * cairo_get_font_face                 (cairo_t *cr);
// void                cairo_set_scaled_font               (cairo_t *cr,
//                                                          const cairo_scaled_font_t *scaled_font);
// cairo_scaled_font_t * cairo_get_scaled_font             (cairo_t *cr);

// void                cairo_show_text                     (cairo_t *cr,
//                                                          const char *utf8);
func (this *Context) ShowText(utf8 string) {
	cutf8 := C.CString(utf8)
	C.cairo_show_text(this.C, cutf8)
	C.free(unsafe.Pointer(cutf8))
}

// void                cairo_show_glyphs                   (cairo_t *cr,
//                                                          const cairo_glyph_t *glyphs,
//                                                          int num_glyphs);
func (this *Context) ShowGlyphs(glyphs []Glyph) {
	var first *C.cairo_glyph_t
	var n = C.int(len(glyphs))
	if n > 0 {
		first = glyphs[0].c()
	}

	C.cairo_show_glyphs(this.C, first, n)
}

// TODO: Implement these
// void                cairo_show_text_glyphs              (cairo_t *cr,
//                                                          const char *utf8,
//                                                          int utf8_len,
//                                                          const cairo_glyph_t *glyphs,
//                                                          int num_glyphs,
//                                                          const cairo_text_cluster_t *clusters,
//                                                          int num_clusters,
//                                                          cairo_text_cluster_flags_t cluster_flags);
// void                cairo_font_extents                  (cairo_t *cr,
//                                                          cairo_font_extents_t *extents);
// void                cairo_text_extents                  (cairo_t *cr,
//                                                          const char *utf8,
//                                                          cairo_text_extents_t *extents);
// void                cairo_glyph_extents                 (cairo_t *cr,
//                                                          const cairo_glyph_t *glyphs,
//                                                          int num_glyphs,
//                                                          cairo_text_extents_t *extents);
// cairo_font_face_t * cairo_toy_font_face_create          (const char *family,
//                                                          cairo_font_slant_t slant,
//                                                          cairo_font_weight_t weight);
// const char *        cairo_toy_font_face_get_family      (cairo_font_face_t *font_face);
// cairo_font_slant_t  cairo_toy_font_face_get_slant       (cairo_font_face_t *font_face);
// cairo_font_weight_t  cairo_toy_font_face_get_weight     (cairo_font_face_t *font_face);
// cairo_glyph_t *     cairo_glyph_allocate                (int num_glyphs);
// void                cairo_glyph_free                    (cairo_glyph_t *glyphs);
// cairo_text_cluster_t * cairo_text_cluster_allocate      (int num_clusters);
// void                cairo_text_cluster_free             (cairo_text_cluster_t *clusters);

//----------------------------------------------------------------------------
// Fonts (TODO)
//----------------------------------------------------------------------------

//----------------------------------------------------------------------------
// Device (TODO)
//----------------------------------------------------------------------------

//----------------------------------------------------------------------------
// Surface
//----------------------------------------------------------------------------

// #define             CAIRO_MIME_TYPE_JP2
// #define             CAIRO_MIME_TYPE_JPEG
// #define             CAIRO_MIME_TYPE_PNG
// #define             CAIRO_MIME_TYPE_URI
const (
	MimeTypeJp2  = "image/jp2"
	MimeTypeJpeg = "image/jpeg"
	MimeTypePng  = "image/png"
	MimeTypeUri  = "text/x-uri"
)

// typedef             cairo_surface_t;
type Surface struct {
	C *C.cairo_surface_t
}

func surface_finalizer(this *Surface) {
	C.cairo_surface_set_user_data(this.C, &go_repr_cookie, nil, nil)
	C.cairo_surface_destroy(this.C)
}

func SurfaceWrap(c *C.cairo_surface_t, grab bool) *Surface {
	go_repr := C.cairo_surface_get_user_data(c, &go_repr_cookie)
	if go_repr != nil {
		return (*Surface)(go_repr)
	}

	surface := &Surface{c}
	if grab {
		C.cairo_pattern_reference(c)
	}
	runtime.SetFinalizer(surface, surface_finalizer)

	status := C.cairo_surface_set_user_data(c, &go_repr_cookie, unsafe.Pointer(surface), nil)
	if status != C.CAIRO_STATUS_SUCCESS {
		panic("failed to set user data, out of memory?")
	}
	return surface

}

// enum                cairo_content_t;
type Content int

const (
	ContentColor      Content = C.CAIRO_CONTENT_COLOR
	ContentAlpha      Content = C.CAIRO_CONTENT_ALPHA
	ContentColorAlpha Content = C.CAIRO_CONTENT_COLOR_ALPHA
)

func (this Content) c() C.cairo_content_t {
	return C.cairo_content_t(this)
}

// cairo_surface_t *   cairo_surface_create_similar        (cairo_surface_t *other,
//                                                          cairo_content_t content,
//                                                          int width,
//                                                          int height);
func (this *Surface) CreateSimilar(content Content, width, height int) *Surface {
	return SurfaceWrap(C.cairo_surface_create_similar(this.C, content.c(), C.int(width), C.int(height)), false)
}

// cairo_surface_t *   cairo_surface_create_for_rectangle  (cairo_surface_t *target,
//                                                          double x,
//                                                          double y,
//                                                          double width,
//                                                          double height);
func (this *Surface) CreateForRectangle(x, y, width, height float64) *Surface {
	return SurfaceWrap(C.cairo_surface_create_for_rectangle(this.C,
		C.double(x), C.double(y), C.double(width), C.double(height)), false)
}

// cairo_surface_t *   cairo_surface_reference             (cairo_surface_t *surface);
// void                cairo_surface_destroy               (cairo_surface_t *surface);

// cairo_status_t      cairo_surface_status                (cairo_surface_t *surface);
func (this *Surface) Status() Status {
	return Status(C.cairo_surface_status(this.C))
}

// void                cairo_surface_finish                (cairo_surface_t *surface);
func (this *Surface) Finish() {
	C.cairo_surface_finish(this.C)
}

// void                cairo_surface_flush                 (cairo_surface_t *surface);
func (this *Surface) Flush() {
	C.cairo_surface_flush(this.C)
}

// TODO: Implement these
// cairo_device_t *    cairo_surface_get_device            (cairo_surface_t *surface);
// void                cairo_surface_get_font_options      (cairo_surface_t *surface,
//                                                          cairo_font_options_t *options);

// cairo_content_t     cairo_surface_get_content           (cairo_surface_t *surface);
func (this *Surface) GetContent() Content {
	return Content(C.cairo_surface_get_content(this.C))
}

// void                cairo_surface_mark_dirty            (cairo_surface_t *surface);
func (this *Surface) MarkDirty() {
	C.cairo_surface_mark_dirty(this.C)
}

// void                cairo_surface_mark_dirty_rectangle  (cairo_surface_t *surface,
//                                                          int x,
//                                                          int y,
//                                                          int width,
//                                                          int height);
func (this *Surface) MarkDirtyRectangle(x, y, width, height int) {
	C.cairo_surface_mark_dirty_rectangle(this.C,
		C.int(x), C.int(y), C.int(width), C.int(height))
}

// void                cairo_surface_set_device_offset     (cairo_surface_t *surface,
//                                                          double x_offset,
//                                                          double y_offset);
func (this *Surface) SetDeviceOffset(x_offset, y_offset float64) {
	C.cairo_surface_set_device_offset(this.C, C.double(x_offset), C.double(y_offset))
}

// void                cairo_surface_get_device_offset     (cairo_surface_t *surface,
//                                                          double *x_offset,
//                                                          double *y_offset);
func (this *Surface) GetDeviceOffset() (x_offset, y_offset float64) {
	C.cairo_surface_get_device_offset(this.C,
		(*C.double)(unsafe.Pointer(&x_offset)),
		(*C.double)(unsafe.Pointer(&y_offset)))
	return
}

// void                cairo_surface_set_fallback_resolution
//                                                         (cairo_surface_t *surface,
//                                                          double x_pixels_per_inch,
//                                                          double y_pixels_per_inch);
func (this *Surface) SetFallbackResolution(x_pixels_per_inch, y_pixels_per_inch float64) {
	C.cairo_surface_set_fallback_resolution(this.C, C.double(x_pixels_per_inch), C.double(y_pixels_per_inch))
}

// void                cairo_surface_get_fallback_resolution
//                                                         (cairo_surface_t *surface,
//                                                          double *x_pixels_per_inch,
//                                                          double *y_pixels_per_inch);
func (this *Surface) GetFallbackResolution() (x_pixels_per_inch, y_pixels_per_inch float64) {
	C.cairo_surface_get_fallback_resolution(this.C,
		(*C.double)(unsafe.Pointer(&x_pixels_per_inch)),
		(*C.double)(unsafe.Pointer(&y_pixels_per_inch)))
	return
}

// enum                cairo_surface_type_t;
type SurfaceType int

const (
	SurfaceTypeImage         SurfaceType = C.CAIRO_SURFACE_TYPE_IMAGE
	SurfaceTypePDF           SurfaceType = C.CAIRO_SURFACE_TYPE_PDF
	SurfaceTypePS            SurfaceType = C.CAIRO_SURFACE_TYPE_PS
	SurfaceTypeXLib          SurfaceType = C.CAIRO_SURFACE_TYPE_XLIB
	SurfaceTypeXCB           SurfaceType = C.CAIRO_SURFACE_TYPE_XCB
	SurfaceTypeGlitz         SurfaceType = C.CAIRO_SURFACE_TYPE_GLITZ
	SurfaceTypeQuartz        SurfaceType = C.CAIRO_SURFACE_TYPE_QUARTZ
	SurfaceTypeWin32         SurfaceType = C.CAIRO_SURFACE_TYPE_WIN32
	SurfaceTypeBeOS          SurfaceType = C.CAIRO_SURFACE_TYPE_BEOS
	SurfaceTypeDirectFB      SurfaceType = C.CAIRO_SURFACE_TYPE_DIRECTFB
	SurfaceTypeSVG           SurfaceType = C.CAIRO_SURFACE_TYPE_SVG
	SurfaceTypeOs2           SurfaceType = C.CAIRO_SURFACE_TYPE_OS2
	SurfaceTypeWin32Printing SurfaceType = C.CAIRO_SURFACE_TYPE_WIN32_PRINTING
	SurfaceTypeQuartzImage   SurfaceType = C.CAIRO_SURFACE_TYPE_QUARTZ_IMAGE
	SurfaceTypeScript        SurfaceType = C.CAIRO_SURFACE_TYPE_SCRIPT
	SurfaceTypeQt            SurfaceType = C.CAIRO_SURFACE_TYPE_QT
	SurfaceTypeRecording     SurfaceType = C.CAIRO_SURFACE_TYPE_RECORDING
	SurfaceTypeVg            SurfaceType = C.CAIRO_SURFACE_TYPE_VG
	SurfaceTypeGL            SurfaceType = C.CAIRO_SURFACE_TYPE_GL
	SurfaceTypeDRM           SurfaceType = C.CAIRO_SURFACE_TYPE_DRM
	SurfaceTypeTee           SurfaceType = C.CAIRO_SURFACE_TYPE_TEE
	SurfaceTypeXML           SurfaceType = C.CAIRO_SURFACE_TYPE_XML
	SurfaceTypeSkia          SurfaceType = C.CAIRO_SURFACE_TYPE_SKIA
	SurfaceTypeSubsurface    SurfaceType = C.CAIRO_SURFACE_TYPE_SUBSURFACE
)

// cairo_surface_type_t  cairo_surface_get_type            (cairo_surface_t *surface);
func (this *Surface) GetType() SurfaceType {
	return SurfaceType(C.cairo_surface_get_type(this.C))
}

// TODO: Implement these
// unsigned int        cairo_surface_get_reference_count   (cairo_surface_t *surface);
// cairo_status_t      cairo_surface_set_user_data         (cairo_surface_t *surface,
//                                                          const cairo_user_data_key_t *key,
//                                                          void *user_data,
//                                                          cairo_destroy_func_t destroy);
// void *              cairo_surface_get_user_data         (cairo_surface_t *surface,
//                                                          const cairo_user_data_key_t *key);

// void                cairo_surface_copy_page             (cairo_surface_t *surface);
func (this *Surface) CopyPage() {
	C.cairo_surface_copy_page(this.C)
}

// void                cairo_surface_show_page             (cairo_surface_t *surface);
func (this *Surface) ShowPage() {
	C.cairo_surface_show_page(this.C)
}

// cairo_bool_t        cairo_surface_has_show_text_glyphs  (cairo_surface_t *surface);
func (this *Surface) HasShowTextGlyphs() bool {
	return C.cairo_surface_has_show_text_glyphs(this.C) != 0
}

// TODO: Implement these
// cairo_status_t      cairo_surface_set_mime_data         (cairo_surface_t *surface,
//                                                          const char *mime_type,
//                                                          unsigned char *data,
//                                                          unsigned long  length,
//                                                          cairo_destroy_func_t destroy,
//                                                          void *closure);
// void                cairo_surface_get_mime_data         (cairo_surface_t *surface,
//                                                          const char *mime_type,
//                                                          unsigned char **data,
//                                                          unsigned long *length);

//----------------------------------------------------------------------------
// Image Surfaces
//----------------------------------------------------------------------------

// enum                cairo_format_t;
type Format int

const (
	FormatInvalid   Format = C.CAIRO_FORMAT_INVALID
	FormatARGB32    Format = C.CAIRO_FORMAT_ARGB32
	FormatRGB24     Format = C.CAIRO_FORMAT_RGB24
	FormatA8        Format = C.CAIRO_FORMAT_A8
	FormatA1        Format = C.CAIRO_FORMAT_A1
	FormatRGB16_565 Format = C.CAIRO_FORMAT_RGB16_565
)

func (this Format) c() C.cairo_format_t {
	return C.cairo_format_t(this)
}

// int                 cairo_format_stride_for_width       (cairo_format_t format,
//                                                          int width);
func (this Format) StrideForWidth(width int) int {
	return int(C.cairo_format_stride_for_width(this.c(), C.int(width)))
}

// cairo_surface_t *   cairo_image_surface_create          (cairo_format_t format,
//                                                          int width,
//                                                          int height);
func NewImageSurface(format Format, width, height int) *Surface {
	return SurfaceWrap(C.cairo_image_surface_create(format.c(), C.int(width), C.int(height)), false)
}

// TODO: Implement this (need a way to keep GC from freeing the 'data')
// cairo_surface_t *   cairo_image_surface_create_for_data (unsigned char *data,
//                                                          cairo_format_t format,
//                                                          int width,
//                                                          int height,
//                                                          int stride);

// TODO: Implement this (think about wrapping it into slice.. how?)
// unsigned char *     cairo_image_surface_get_data        (cairo_surface_t *surface);

// cairo_format_t      cairo_image_surface_get_format      (cairo_surface_t *surface);
func (this *Surface) GetFormat() Format {
	return Format(C.cairo_image_surface_get_format(this.C))
}

// int                 cairo_image_surface_get_width       (cairo_surface_t *surface);
func (this *Surface) GetWidth() int {
	return int(C.cairo_image_surface_get_width(this.C))
}

// int                 cairo_image_surface_get_height      (cairo_surface_t *surface);
func (this *Surface) GetHeight() int {
	return int(C.cairo_image_surface_get_height(this.C))
}

// int                 cairo_image_surface_get_stride      (cairo_surface_t *surface);
func (this *Surface) GetStride() int {
	return int(C.cairo_image_surface_get_stride(this.C))
}

//----------------------------------------------------------------------------
// PNG Support
//----------------------------------------------------------------------------

// cairo_surface_t *   cairo_image_surface_create_from_png (const char *filename);
func NewImageSurfaceFromPNG(filename string) *Surface {
	cfilename := C.CString(filename)
	surface := SurfaceWrap(C.cairo_image_surface_create_from_png(cfilename), false)
	C.free(unsafe.Pointer(cfilename))
	return surface
}

// cairo_status_t      (*cairo_read_func_t)                (void *closure,
//                                                          unsigned char *data,
//                                                          unsigned int length);

//export io_reader_wrapper
func io_reader_wrapper(reader_up unsafe.Pointer, data_up unsafe.Pointer, length uint32) uint32 {
	var reader io.Reader
	var data []byte
	var data_header reflect.SliceHeader
	reader = *(*io.Reader)(reader_up)
	data_header.Data = uintptr(data_up)
	data_header.Len = int(length)
	data_header.Cap = int(length)
	data = *(*[]byte)(unsafe.Pointer(&data_header))

	_, err := reader.Read(data)
	if err != nil {
		return uint32(StatusReadError)
	}
	return uint32(StatusSuccess)
}

// cairo_surface_t *   cairo_image_surface_create_from_png_stream
//                                                         (cairo_read_func_t read_func,
//                                                          void *closure);
func NewImageSurfaceFromPNGStream(r io.Reader) *Surface {
	return SurfaceWrap(C._cairo_image_surface_create_from_png_stream(unsafe.Pointer(&r)), false)
}

// cairo_status_t      cairo_surface_write_to_png          (cairo_surface_t *surface,
//                                                          const char *filename);
func (this *Surface) WriteToPNG(filename string) Status {
	cfilename := C.CString(filename)
	status := C.cairo_surface_write_to_png(this.C, cfilename)
	C.free(unsafe.Pointer(cfilename))
	return Status(status)
}

// cairo_status_t      (*cairo_write_func_t)               (void *closure,
//                                                          unsigned char *data,
//                                                          unsigned int length);

//export io_writer_wrapper
func io_writer_wrapper(writer_up unsafe.Pointer, data_up unsafe.Pointer, length uint32) uint32 {
	var writer io.Writer
	var data []byte
	var data_header reflect.SliceHeader
	writer = *(*io.Writer)(writer_up)
	data_header.Data = uintptr(data_up)
	data_header.Len = int(length)
	data_header.Cap = int(length)
	data = *(*[]byte)(unsafe.Pointer(&data_header))

	_, err := writer.Write(data)
	if err != nil {
		return uint32(StatusWriteError)
	}
	return uint32(StatusSuccess)
}

// cairo_status_t      cairo_surface_write_to_png_stream   (cairo_surface_t *surface,
//                                                          cairo_write_func_t write_func,
//                                                          void *closure);
func (this *Surface) WriteToPNGStream(w io.Writer) Status {
	return Status(C._cairo_surface_write_to_png_stream(this.C, unsafe.Pointer(&w)))
}

//----------------------------------------------------------------------------
// Matrix
//----------------------------------------------------------------------------

//                     cairo_matrix_t;
type Matrix struct {
	xx, yx, xy, yy, x0, y0 float64
}

func (this *Matrix) c() *C.cairo_matrix_t {
	return (*C.cairo_matrix_t)(unsafe.Pointer(this))
}

// void                cairo_matrix_init                   (cairo_matrix_t *matrix,
//                                                          double xx,
//                                                          double yx,
//                                                          double xy,
//                                                          double yy,
//                                                          double x0,
//                                                          double y0);
func (this *Matrix) Init(xx, yx, xy, yy, x0, y0 float64) {
	C.cairo_matrix_init(this.c(), C.double(xx), C.double(yx), C.double(xy), C.double(yy), C.double(x0), C.double(y0))
}

// void                cairo_matrix_init_identity          (cairo_matrix_t *matrix);
func (this *Matrix) InitIdentity() {
	C.cairo_matrix_init_identity(this.c())
}

// void                cairo_matrix_init_translate         (cairo_matrix_t *matrix,
//                                                          double tx,
//                                                          double ty);
func (this *Matrix) InitTranslate(tx, ty float64) {
	C.cairo_matrix_init_translate(this.c(), C.double(tx), C.double(ty))
}

// void                cairo_matrix_init_scale             (cairo_matrix_t *matrix,
//                                                          double sx,
//                                                          double sy);
func (this *Matrix) InitScale(sx, sy float64) {
	C.cairo_matrix_init_scale(this.c(), C.double(sx), C.double(sy))
}

// void                cairo_matrix_init_rotate            (cairo_matrix_t *matrix,
//                                                          double radians);
func (this *Matrix) InitRotate(radians float64) {
	C.cairo_matrix_init_rotate(this.c(), C.double(radians))
}

// void                cairo_matrix_translate              (cairo_matrix_t *matrix,
//                                                          double tx,
//                                                          double ty);
func (this *Matrix) Translate(tx, ty float64) {
	C.cairo_matrix_translate(this.c(), C.double(tx), C.double(ty))
}

// void                cairo_matrix_scale                  (cairo_matrix_t *matrix,
//                                                          double sx,
//                                                          double sy);
func (this *Matrix) Scale(sx, sy float64) {
	C.cairo_matrix_scale(this.c(), C.double(sx), C.double(sy))
}

// void                cairo_matrix_rotate                 (cairo_matrix_t *matrix,
//                                                          double radians);
func (this *Matrix) Rotate(radians float64) {
	C.cairo_matrix_rotate(this.c(), C.double(radians))
}

// cairo_status_t      cairo_matrix_invert                 (cairo_matrix_t *matrix);
func (this *Matrix) Invert() Status {
	return Status(C.cairo_matrix_invert(this.c()))
}

// void                cairo_matrix_multiply               (cairo_matrix_t *result,
//                                                          const cairo_matrix_t *a,
//                                                          const cairo_matrix_t *b);
func (this *Matrix) Multiply(b *Matrix) Matrix {
	var result C.cairo_matrix_t
	C.cairo_matrix_multiply(&result, this.c(), b.c())
	return *(*Matrix)(unsafe.Pointer(&result))
}

// void                cairo_matrix_transform_distance     (const cairo_matrix_t *matrix,
//                                                          double *dx,
//                                                          double *dy);
func (this *Matrix) TransformDistance(dx, dy float64) (float64, float64) {
	C.cairo_matrix_transform_distance(this.c(),
		(*C.double)(unsafe.Pointer(&dx)),
		(*C.double)(unsafe.Pointer(&dy)))
	return dx, dy
}

// void                cairo_matrix_transform_point        (const cairo_matrix_t *matrix,
//                                                          double *x,
//                                                          double *y);
func (this *Matrix) TransformPoint(x, y float64) (float64, float64) {
	C.cairo_matrix_transform_point(this.c(),
		(*C.double)(unsafe.Pointer(&x)),
		(*C.double)(unsafe.Pointer(&y)))
	return x, y
}
