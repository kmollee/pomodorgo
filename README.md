# Pomodorogo

> "Pomodoro" is a time management method

![demo](images/pomodorogo.gif)

Pomodorogo: A configurable pomodoro timer write in go

inspired by [countdown](https://github.com/antonmedv/countdown)

## Usage

to run pomodorogo timer(first time launch will automatic create config for you at `~/pomodorgo.ini`)

```sh
pomodorogo
```

with custom config file path

```sh
pomodorogo -c ./myconfig.ini
```


Pomodorogo run with config file `pomodorogo.ini`.

Config file must have `schedule` field to define what sections going to run

Each section must specific time, and section can define `cmd`, that will run when section is start and terminate when end

below is an example config file

```ini
# schedule contain list of section
schedule = section short-break section short-break section long-break

[section]
time = 25m
# cmd = mpv --loop woogie_boogie_pinao.mp3

[short-break]
time = 5m

[long-break]
time = 15m
```


# Key binding

- p or P: To pause the timer.
- c or C: To resume the timer.
- Esc or Ctrl+C or q or Q: To stop the timer.
- n or N: To jump next schedule.

# Install

```
go get github.com/kmollee/pomodorogo
```

# TODO

- [ ] examine config file correct

