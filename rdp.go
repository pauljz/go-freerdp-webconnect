package main

/*
#cgo LDFLAGS: -L. -lfreerdp-core -lfreerdp-codec -lfreerdp-gdi
#include <freerdp/freerdp.h>
#include <freerdp/codec/color.h>
#include <freerdp/gdi/gdi.h>
#include <unistd.h>

extern boolean preConnect(freerdp* instance);
extern void postConnect(freerdp* instance);
extern void goPrintln(char* text);
extern void goEcho(char* text, rdpContext* context);
extern size_t getPointerSize();
extern void primaryPatBlt(rdpContext* context, PATBLT_ORDER* patblt);
extern void primaryScrBlt(rdpContext* context, SCRBLT_ORDER* scrblt);
extern void primaryOpaqueRect(rdpContext* context, OPAQUE_RECT_ORDER* oro);
extern void primaryMultiOpaqueRect(rdpContext* context, MULTI_OPAQUE_RECT_ORDER* moro);
extern void beginPaint(rdpContext* context);
extern void endPaint(rdpContext* context);
extern void setBounds(rdpContext* context, rdpBounds* bounds);
extern void bitmapUpdate(rdpContext* context, BITMAP_UPDATE* bitmap);

static void cbPrimaryPatBlt(rdpContext* context, PATBLT_ORDER* patblt) {
	primaryPatBlt(context, patblt);
}

static void cbPrimaryScrBlt(rdpContext* context, SCRBLT_ORDER* scrblt) {
	primaryScrBlt(context, scrblt);
}

static void cbPrimaryOpaqueRect(rdpContext* context, OPAQUE_RECT_ORDER* oro) {
	primaryOpaqueRect(context, oro);
}

static void cbPrimaryMultiOpaqueRect(rdpContext* context, MULTI_OPAQUE_RECT_ORDER* moro) {
	primaryMultiOpaqueRect(context, moro);
}

static void cbBeginPaint(rdpContext* context) {
	//beginPaint(context);
}
static void cbEndPaint(rdpContext* context) {
	//endPaint(context);
}
static void cbSetBounds(rdpContext* context, rdpBounds* bounds) {
	//setBounds(context, bounds);
}
static void cbBitmapUpdate(rdpContext* context, BITMAP_UPDATE* bitmap) {
	bitmapUpdate(context, bitmap);
}

static boolean cbPreConnect(freerdp* instance) {
	rdpUpdate* update = instance->update;
	rdpPrimaryUpdate* primary = update->primary;

	primary->PatBlt = cbPrimaryPatBlt;
	primary->ScrBlt = cbPrimaryScrBlt;
	primary->OpaqueRect = cbPrimaryOpaqueRect;
	primary->MultiOpaqueRect = cbPrimaryMultiOpaqueRect;

	update->BeginPaint = cbBeginPaint;
	update->EndPaint = cbEndPaint;
	update->SetBounds = cbSetBounds;
	update->BitmapUpdate = cbBitmapUpdate;

	return preConnect(instance);
}

static boolean cbPostConnect(freerdp* instance) {
	postConnect(instance);

	rdpPointer p;
	memset(&p, 0, sizeof(p));

	p.size = getPointerSize();

	//p.New = cbPointer_New;
	//p.Free = cbPointer_Free;
	//p.Set = cbPointer_Set;
	//p.SetNull = cbPointer_SetNull;
	//p.SetDefault = cbPointer_SetDefault;

	graphics_register_pointer(instance->context->graphics, &p);

	return 1;
}

static BITMAP_DATA* nextBitmapRectangle(BITMAP_UPDATE* bitmap, int i) {
	return &bitmap->rectangles[i];
}

static DELTA_RECT* nextMultiOpaqueRectangle(MULTI_OPAQUE_RECT_ORDER* moro, int i) {
	return &moro->rectangles[i];
}

static void bindCallbacks(freerdp* instance) {
	instance->PreConnect = cbPreConnect;
	instance->PostConnect = cbPostConnect;
}
*/
import (
	"C"
)
import (
	"bytes"
	"encoding/binary"
	"fmt"
	"reflect"
	"unsafe"
)

const (
	WSOP_SC_BEGINPAINT       uint32 = 0
	WSOP_SC_ENDPAINT         uint32 = 1
	WSOP_SC_BITMAP           uint32 = 2
	WSOP_SC_OPAQUERECT       uint32 = 3
	WSOP_SC_SETBOUNDS        uint32 = 4
	WSOP_SC_PATBLT           uint32 = 5
	WSOP_SC_MULTI_OPAQUERECT uint32 = 6
	WSOP_SC_SCRBLT           uint32 = 7
	WSOP_SC_PTR_NEW          uint32 = 8
	WSOP_SC_PTR_FREE         uint32 = 9
	WSOP_SC_PTR_SET          uint32 = 10
	WSOP_SC_PTR_SETNULL      uint32 = 11
	WSOP_SC_PTR_SETDEFAULT   uint32 = 12
)

type bitmapUpdateMeta struct {
	op  uint32
	x   uint32
	y   uint32
	w   uint32
	h   uint32
	dw  uint32
	dh  uint32
	bpp uint32
	cf  uint32
	sz  uint32
}

type primaryPatBltMeta struct {
	op  uint32
	x   int32
	y   int32
	w   int32
	h   int32
	fg  uint32
	rop uint32
}

type primaryScrBltMeta struct {
	op  uint32
	rop uint32
	x   int32
	y   int32
	w   int32
	h   int32
	sx  int32
	sy  int32
}

type rdpConnectionSettings struct {
	hostname *string
	username *string
	password *string
	width    int
	height   int
}

type rdpContext struct {
	_p       C.rdpContext
	clrconv  C.HCLRCONV
	recvq    chan []byte
	sendq    chan []byte
	settings *rdpConnectionSettings
}

type rdpPointer struct {
	pointer *C.rdpPointer
	id      int
}

func rdpconnect(sendq chan []byte, recvq chan []byte, settings *rdpConnectionSettings) {
	var instance *C.freerdp

	fmt.Println("RDP Connecting...")

	instance = C.freerdp_new()

	C.bindCallbacks(instance)
	instance.context_size = C.size_t(unsafe.Sizeof(rdpContext{}))
	C.freerdp_context_new(instance)

	var context *rdpContext
	context = (*rdpContext)(unsafe.Pointer(instance.context))
	context.sendq = sendq
	context.recvq = recvq
	context.settings = settings

	C.freerdp_connect(instance)

	mainEventLoop := true

	for mainEventLoop {
		select {
		case <- recvq:
			fmt.Println("Disconnecting (websocket error)")
			mainEventLoop = false
		default:
			e := int(C.freerdp_error_info(instance))
			if e != 0 {
				switch e {
				case 1:
				case 2:
				case 7:
				case 9:
					// Manual disconnections and such
					fmt.Println("Disconnecting (manual)")
					mainEventLoop = false
					break
				case 5:
					// Another user connected
					break
				default:
					// Unknown error?
					break
				}
			}
			if int(C.freerdp_shall_disconnect(instance)) != 0 {
				fmt.Println("Disconnecting (RDC said so)")
				mainEventLoop = false
			}
			if mainEventLoop {
				C.freerdp_check_fds(instance)
			}
			C.usleep(1000)
		}
	}
	C.freerdp_free(instance)
}

func sendBinary(sendq chan []byte, data *bytes.Buffer) {
	sendq <- data.Bytes()
}

//export getPointerSize
func getPointerSize() C.size_t {
	return C.size_t(unsafe.Sizeof(rdpPointer{}))
}

//export primaryPatBlt
func primaryPatBlt(rawContext *C.rdpContext, patblt *C.PATBLT_ORDER) {
	context := (*rdpContext)(unsafe.Pointer(rawContext))

	hclrconv := context.clrconv

	if C.GDI_BS_SOLID == patblt.brush.style {
		meta := primaryPatBltMeta{
			WSOP_SC_PATBLT,
			int32(patblt.nLeftRect),
			int32(patblt.nTopRect),
			int32(patblt.nWidth),
			int32(patblt.nHeight),
			uint32(C.freerdp_color_convert_var(patblt.foreColor, 16, 32, hclrconv)),
			uint32(C.gdi_rop3_code(C.uint8(patblt.bRop))),
		}

		buf := new(bytes.Buffer)
		binary.Write(buf, binary.LittleEndian, meta)
		sendBinary(context.sendq, buf)
	}
}

//export primaryScrBlt
func primaryScrBlt(rawContext *C.rdpContext, scrblt *C.SCRBLT_ORDER) {
	context := (*rdpContext)(unsafe.Pointer(rawContext))

	meta := primaryScrBltMeta{
		WSOP_SC_SCRBLT,
		uint32(C.gdi_rop3_code(C.uint8(scrblt.bRop))),
		int32(scrblt.nLeftRect),
		int32(scrblt.nTopRect),
		int32(scrblt.nWidth),
		int32(scrblt.nHeight),
		int32(scrblt.nXSrc),
		int32(scrblt.nYSrc),
	}

	buf := new(bytes.Buffer)
	binary.Write(buf, binary.LittleEndian, meta)
	sendBinary(context.sendq, buf)
}

//export primaryOpaqueRect
func primaryOpaqueRect(rawContext *C.rdpContext, oro *C.OPAQUE_RECT_ORDER) {
	context := (*rdpContext)(unsafe.Pointer(rawContext))

	hclrconv := context.clrconv
	svcolor := oro.color
	oro.color = C.freerdp_color_convert_var(oro.color, 16, 32, hclrconv)

	buf := new(bytes.Buffer)
	binary.Write(buf, binary.LittleEndian, WSOP_SC_OPAQUERECT)
	binary.Write(buf, binary.LittleEndian, oro)
	sendBinary(context.sendq, buf)

	oro.color = svcolor
}

//export primaryMultiOpaqueRect
func primaryMultiOpaqueRect(rawContext *C.rdpContext, moro *C.MULTI_OPAQUE_RECT_ORDER) {
	context := (*rdpContext)(unsafe.Pointer(rawContext))

	hclrconv := context.clrconv
	color := C.freerdp_color_convert_var(moro.color, 16, 32, hclrconv)

	buf := new(bytes.Buffer)
	binary.Write(buf, binary.LittleEndian, WSOP_SC_MULTI_OPAQUERECT)
	binary.Write(buf, binary.LittleEndian, int32(color))
	binary.Write(buf, binary.LittleEndian, int32(moro.numRectangles))

	var r *C.DELTA_RECT
	var i int
	for i = 1; i <= int(moro.numRectangles); i++ {
		r = C.nextMultiOpaqueRectangle(moro, C.int(i))
		binary.Write(buf, binary.LittleEndian, r)
	}

	sendBinary(context.sendq, buf)
}

//export beginPaint
func beginPaint(rawContext *C.rdpContext) {
	context := (*rdpContext)(unsafe.Pointer(rawContext))
	buf := new(bytes.Buffer)
	binary.Write(buf, binary.LittleEndian, WSOP_SC_BEGINPAINT)
	sendBinary(context.sendq, buf)
}

//export endPaint
func endPaint(rawContext *C.rdpContext) {
	context := (*rdpContext)(unsafe.Pointer(rawContext))
	buf := new(bytes.Buffer)
	binary.Write(buf, binary.LittleEndian, WSOP_SC_ENDPAINT)
	sendBinary(context.sendq, buf)
}

//export setBounds
func setBounds(rawContext *C.rdpContext, bounds *C.rdpBounds) {
	context := (*rdpContext)(unsafe.Pointer(rawContext))
	buf := new(bytes.Buffer)

	if bounds != nil {
		binary.Write(buf, binary.LittleEndian, WSOP_SC_SETBOUNDS)
		binary.Write(buf, binary.LittleEndian, bounds)
		sendBinary(context.sendq, buf)
	}
}

//export bitmapUpdate
func bitmapUpdate(rawContext *C.rdpContext, bitmap *C.BITMAP_UPDATE) {
	context := (*rdpContext)(unsafe.Pointer(rawContext))

	var bmd *C.BITMAP_DATA
	var i int

	for i = 0; i < int(bitmap.number); i++ {
		bmd = C.nextBitmapRectangle(bitmap, C.int(i))

		buf := new(bytes.Buffer)

		meta := bitmapUpdateMeta{
			WSOP_SC_BITMAP,                           // op
			uint32(bmd.destLeft),                     // x
			uint32(bmd.destTop),                      // y
			uint32(bmd.width),                        // w
			uint32(bmd.height),                       // h
			uint32(bmd.destRight - bmd.destLeft + 1), // dw
			uint32(bmd.destBottom - bmd.destTop + 1), // dh
			uint32(bmd.bitsPerPixel),                 // bpp
			uint32(bmd.compressed),                   // cf
			uint32(bmd.bitmapLength),                 // sz
		}
		if int(bmd.compressed) == 0 {
			C.freerdp_image_flip(bmd.bitmapDataStream, bmd.bitmapDataStream, C.int(bmd.width), C.int(bmd.height), C.int(bmd.bitsPerPixel))
		}

		binary.Write(buf, binary.LittleEndian, meta)

		// Unsafe copy bmd.bitmapLength bytes out of bmd.bitmapDataStream
		var bitmapDataStream []byte
		clen := int(bmd.bitmapLength)
		bitmapDataStream = (*[1 << 30]byte)(unsafe.Pointer(bmd.bitmapDataStream))[:clen]
		(*reflect.SliceHeader)(unsafe.Pointer(&bitmapDataStream)).Cap = clen
		binary.Write(buf, binary.LittleEndian, bitmapDataStream)

		sendBinary(context.sendq, buf)
	}
}

//export postConnect
func postConnect(instance *C.freerdp) {
	fmt.Println("Connected.")
}

//export preConnect
func preConnect(instance *C.freerdp) C.boolean {
	settings := instance.settings
	context := (*rdpContext)(unsafe.Pointer(instance.context))

	settings.hostname = C.CString(*context.settings.hostname)
	settings.username = C.CString(*context.settings.username)
	settings.password = C.CString(*context.settings.password)
	settings.width = C.uint32(context.settings.width)
	settings.height = C.uint32(context.settings.height)

	settings.port = C.uint32(3389)
	settings.ignore_certificate = C.boolean(1)

	settings.performance_flags =
		C.PERF_DISABLE_WALLPAPER |
		C.PERF_DISABLE_THEMING |
		C.PERF_DISABLE_MENUANIMATIONS |
		C.PERF_DISABLE_FULLWINDOWDRAG

	settings.connection_type = C.CONNECTION_TYPE_BROADBAND_HIGH

	settings.rfx_codec = C.boolean(0)
	settings.fastpath_output = C.boolean(1)
	settings.color_depth = C.uint32(16)
	settings.frame_acknowledge = C.boolean(1)
	settings.large_pointer = C.boolean(1)
	settings.glyph_cache = C.boolean(0)
	settings.bitmap_cache = C.boolean(0)
	settings.offscreen_bitmap_cache = C.boolean(0)

	settings.order_support[C.NEG_DSTBLT_INDEX] = 0
	settings.order_support[C.NEG_PATBLT_INDEX] = 1
	settings.order_support[C.NEG_SCRBLT_INDEX] = 1
	settings.order_support[C.NEG_MEMBLT_INDEX] = 0
	settings.order_support[C.NEG_MEM3BLT_INDEX] = 0
	settings.order_support[C.NEG_ATEXTOUT_INDEX] = 0
	settings.order_support[C.NEG_AEXTTEXTOUT_INDEX] = 0
	settings.order_support[C.NEG_DRAWNINEGRID_INDEX] = 0
	settings.order_support[C.NEG_LINETO_INDEX] = 0
	settings.order_support[C.NEG_MULTI_DRAWNINEGRID_INDEX] = 0
	settings.order_support[C.NEG_OPAQUE_RECT_INDEX] = 1
	settings.order_support[C.NEG_SAVEBITMAP_INDEX] = 0
	settings.order_support[C.NEG_WTEXTOUT_INDEX] = 0
	settings.order_support[C.NEG_MEMBLT_V2_INDEX] = 0
	settings.order_support[C.NEG_MEM3BLT_V2_INDEX] = 0
	settings.order_support[C.NEG_MULTIDSTBLT_INDEX] = 0
	settings.order_support[C.NEG_MULTIPATBLT_INDEX] = 0
	settings.order_support[C.NEG_MULTISCRBLT_INDEX] = 0
	settings.order_support[C.NEG_MULTIOPAQUERECT_INDEX] = 1
	settings.order_support[C.NEG_FAST_INDEX_INDEX] = 0
	settings.order_support[C.NEG_POLYGON_SC_INDEX] = 0
	settings.order_support[C.NEG_POLYGON_CB_INDEX] = 0
	settings.order_support[C.NEG_POLYLINE_INDEX] = 0
	settings.order_support[C.NEG_FAST_GLYPH_INDEX] = 0
	settings.order_support[C.NEG_ELLIPSE_SC_INDEX] = 0
	settings.order_support[C.NEG_ELLIPSE_CB_INDEX] = 0
	settings.order_support[C.NEG_GLYPH_INDEX_INDEX] = 0
	settings.order_support[C.NEG_GLYPH_WEXTTEXTOUT_INDEX] = 0
	settings.order_support[C.NEG_GLYPH_WLONGTEXTOUT_INDEX] = 0
	settings.order_support[C.NEG_GLYPH_WLONGEXTTEXTOUT_INDEX] = 0

	context.clrconv = C.freerdp_clrconv_new(C.CLRCONV_ALPHA | C.CLRCONV_INVERT)

	return 1
}
