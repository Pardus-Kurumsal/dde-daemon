/**
 * Copyright (C) 2014 Deepin Technology Co., Ltd.
 *
 * This program is free software; you can redistribute it and/or modify
 * it under the terms of the GNU General Public License as published by
 * the Free Software Foundation; either version 3 of the License, or
 * (at your option) any later version.
 **/

#include <gtk/gtk.h>

#include "cursor.h"

static gboolean end_flag = FALSE;
static void update_gtk_cursor();

void
handle_gtk_cursor_changed()
{
    static int sig_id = 0;
    if (sig_id > 0) {
        end_flag = FALSE;
        g_debug("Cursor changed handler has running\n");
        return;
    }

    gtk_init(NULL, NULL);
    GtkSettings* s = gtk_settings_get_default();
    sig_id = g_signal_connect(s, "notify::gtk-cursor-theme-name",
                              update_gtk_cursor, NULL);
    if (sig_id <= 0) {
        g_warning("Connect gtk cursor changed failed!");
        return;
    }

    gtk_main();
}

void
end_cursor_changed_handler()
{
    end_flag = TRUE;
}

static void
update_gtk_cursor()
{
    if (end_flag) {
        return ;
    }

    GdkCursor* cursor = gdk_cursor_new_for_display(
        gdk_display_get_default(),
        GDK_LEFT_PTR);
    gdk_window_set_cursor(gdk_get_default_root_window(), cursor);
    g_object_unref(G_OBJECT(cursor));
}
