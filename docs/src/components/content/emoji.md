# Emoji Shortcodes

Anywhere text is decoration rather than data, a `:name:` shortcode expands to
the emoji it stands for, the same names GitHub and Slack use.

```go
tgcomp.Text(p.Main, "Shipped it :tada:")
```

Renders as `Shipped it 🎉`.

## Where it applies

| Expands | Stays literal |
| --- | --- |
| `Title`, `Subtitle`, `Text` | `Code` |
| `Link` text, but not its url | code spans and fenced blocks in `Markdown` |
| prose in `Markdown` | `Html`, `Json`, `Table` |
| `PageConfig.Emoji` | anything a user typed and the app reads back |

A shortcode inside code is the thing being shown, not decoration, so it is
left as written:

```go
tgcomp.Markdown(p.Main, "A `:tada:` in code stays as written.")
```

That is also the way to show a shortcode literally — there is no escape
character.

## Names

A name the table does not know stays as it was written, so `:notanemoji:`
renders as typed rather than disappearing. The names come from
[emojibase](https://emojibase.dev/docs/datasets/)'s github preset, 1913 of
them; `:tada:`, `:+1:`, `:100:` and the rest of what you would type in a
pull request.

Only whole shortcodes are replaced, so ordinary text that happens to hold
colons is safe: a timestamp like `15:04:05` comes through untouched, because
`04` is not a name.
