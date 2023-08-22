# stitch

What happens if:
- someone uses template values in an unapproved field
    - template values remain literal, other, approprirate values are properly templated
- rules and actions in the same dir
    - gets loaded as a template, name only as no other fields will map
- user-provided dir does not exist
    - correctly panics and cites full dir path
- template data is missing
    - GO subs <no value> in place of template identifier
- request method validation
- slack error handling
    - test with invalid token
        - logs auth issue, returns 500. This should bubble up auth issue or return an error message?
    - test with private channel
        - returns `channel_not_found` & 500. Again, bubble up
- defer