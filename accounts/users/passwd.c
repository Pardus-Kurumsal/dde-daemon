/**
 * Copyright (C) 2013 Deepin Technology Co., Ltd.
 *
 * This program is free software; you can redistribute it and/or modify
 * it under the terms of the GNU General Public License as published by
 * the Free Software Foundation; either version 3 of the License, or
 * (at your option) any later version.
 **/

#include <time.h>
#include <unistd.h>
#include <crypt.h>
#include <shadow.h>

#include "passwd.h"

char *
mkpasswd (const char *words)
{
    unsigned long seed[2];
    char salt[] = "$6$........";
    const char *const seedchars =
        "./0123456789ABCDEFGHIJKLMNOPQRST"
        "UVWXYZabcdefghijklmnopqrstuvwxyz";
    char *password;
    int i;

    // Generate a (not very) random seed. You should do it better than this...
    seed[0] = time(NULL);
    seed[1] = getpid() ^ (seed[0] >> 14 & 0x30000);

    // Turn it into printable characters from `seedchars'.
    for (i = 0; i < 8; i++) {
        salt[3 + i] = seedchars[(seed[i / 5] >> (i % 5) * 6) & 0x3f];
    }

    // DES Encrypt
    password = crypt(words, salt);

    return password;
}

int
lock_shadow_file()
{
	return lckpwdf();
}

int
unlock_shadow_file()
{
	return ulckpwdf();
}
