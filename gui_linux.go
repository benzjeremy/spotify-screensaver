//go:build linux && cgo

package main

/*
#cgo pkg-config: webkit2gtk-4.1 gtk+-3.0
#include <stdlib.h>
#include <gtk/gtk.h>
#include <webkit2/webkit2.h>

static int check_display() {
    int argc = 0;
    char **argv = NULL;
    return gtk_init_check(&argc, &argv) ? 1 : 0;
}

static void on_window_destroy(GtkWidget *widget, gpointer data) {
    gtk_main_quit();
}

static gboolean on_context_menu(WebKitWebView *web_view, WebKitContextMenu *context_menu, GdkEvent *event, WebKitHitTestResult *hit_test_result, gpointer user_data) {
    return TRUE; // Suppress browser context menu for native screensaver feeling
}

static gboolean on_key_press(GtkWidget *widget, GdkEventKey *event, gpointer user_data) {
    if (event->keyval == GDK_KEY_Escape) {
        gtk_main_quit();
        return TRUE;
    }
    if (event->keyval == GDK_KEY_F11) {
        static gboolean is_fullscreen = FALSE;
        if (is_fullscreen) {
            gtk_window_unfullscreen(GTK_WINDOW(widget));
            is_fullscreen = FALSE;
        } else {
            gtk_window_fullscreen(GTK_WINDOW(widget));
            is_fullscreen = TRUE;
        }
        return TRUE;
    }
    return FALSE;
}

static void run_gtk_screensaver(const char *title, const char *url, int width, int height, int fullscreen) {
    int argc = 0;
    char **argv = NULL;
    if (!gtk_init_check(&argc, &argv)) {
        return;
    }

    g_set_prgname("spotify-screensaver");
    g_set_application_name("Spotify Screensaver");

    GtkWidget *window = gtk_window_new(GTK_WINDOW_TOPLEVEL);
    gtk_window_set_title(GTK_WINDOW(window), title);
    gtk_window_set_default_size(GTK_WINDOW(window), width, height);
    gtk_window_set_position(GTK_WINDOW(window), GTK_WIN_POS_CENTER);

    GdkRGBA bg_color;
    gdk_rgba_parse(&bg_color, "#07090e");
    #pragma GCC diagnostic push
    #pragma GCC diagnostic ignored "-Wdeprecated-declarations"
    gtk_widget_override_background_color(window, GTK_STATE_FLAG_NORMAL, &bg_color);
    #pragma GCC diagnostic pop

    WebKitSettings *settings = webkit_settings_new();
    webkit_settings_set_enable_developer_extras(settings, FALSE);
    webkit_settings_set_enable_javascript(settings, TRUE);
    webkit_settings_set_enable_webaudio(settings, TRUE);
    webkit_settings_set_hardware_acceleration_policy(settings, WEBKIT_HARDWARE_ACCELERATION_POLICY_ALWAYS);

    GtkWidget *web_view = webkit_web_view_new_with_settings(settings);
    g_signal_connect(web_view, "context-menu", G_CALLBACK(on_context_menu), NULL);
    g_signal_connect(window, "key-press-event", G_CALLBACK(on_key_press), NULL);

    gtk_container_add(GTK_CONTAINER(window), web_view);
    g_signal_connect(window, "destroy", G_CALLBACK(on_window_destroy), NULL);

    webkit_web_view_load_uri(WEBKIT_WEB_VIEW(web_view), url);
    gtk_widget_show_all(window);

    if (fullscreen) {
        gtk_window_fullscreen(GTK_WINDOW(window));
    }

    gtk_main();
}
*/
import "C"
import (
	"log"
	"os/exec"
	"unsafe"
)

func LaunchGUI(title, url string, width, height int, fullscreen bool) {
	if C.check_display() == 0 {
		log.Println("[GUI] Kein X11/Wayland Display gefunden, öffne Standardbrowser...")
		_ = exec.Command("xdg-open", url).Start()
		return
	}

	cTitle := C.CString(title)
	cURL := C.CString(url)
	defer C.free(unsafe.Pointer(cTitle))
	defer C.free(unsafe.Pointer(cURL))

	cFullscreen := C.int(0)
	if fullscreen {
		cFullscreen = C.int(1)
	}

	log.Printf("[GUI] Starte WebKitGTK Screensaver (%s)...\n", url)
	C.run_gtk_screensaver(cTitle, cURL, C.int(width), C.int(height), cFullscreen)
}
