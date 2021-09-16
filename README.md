# jl - json log viewer

a drop in replacement for `less`, with some different features

It is meant to read logs, and parse them line-by-line.

will try to parse each line as json, expecting this fields:

    {
        "message": ...,
        "time": ...,
        "level": ...,
        ...,
    }

only a compact time/level is shown, and the message. everything else is hidden in the normal view but can be
searched ('/') or viewed by switching level of details (press 'D')

you can then search over the lines, and use the search results to make in-memory buffers to further search on

press `h` for in-app help
