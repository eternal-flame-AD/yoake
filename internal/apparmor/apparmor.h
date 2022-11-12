#include <stdio.h>
#include <stdlib.h>
#include <errno.h>
#include <sys/apparmor.h>

int go_aa_change_hat(const char *hat, unsigned long magic);

int go_aa_change_profile(const char *profile);