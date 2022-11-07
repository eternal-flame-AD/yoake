#!/usr/bin/env fish

find . -executable -regextype awk -iregex '.+\.build\.(sh|bash|fish)'
for script in \
    (find . -executable -regextype awk -iregex '.+\.build.(sh|bash|fish)')

    set -l script_path (realpath $script)
    set -l script_dir (dirname $script_path)
    cd $script_dir
    echo "--> Build \$WEBROOT/$script_path"
        $script_path
        or begin
            set -l exit_code $status
            echo "---> : Command $script_path returned $exit_code"
            exit $exit_code
        end
    cd -
end

find . -executable -regextype awk -iregex '.+\.build.(sh|bash|fish)' -delete