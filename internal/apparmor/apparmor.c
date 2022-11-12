#include "./apparmor.h"

int go_aa_change_hat(const char *hat, unsigned long magic)
{
    int ret = aa_change_hat(hat, magic);
    if (ret < 0)
    {
        return errno;
    }
    return 0;
}

int go_aa_change_profile(const char *profile)
{
    int ret = aa_change_profile(profile);
    if (ret < 0)
    {
        return errno;
    }
    return 0;
}