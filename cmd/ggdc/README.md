# GGDC - General Game Description Compiler

At the moment, this only emits the tokenized/parsed version of a GDL (HRF) input
from stdin to stdout.  The goal is for this to be able to translate from the GDL
into one of [ a service definition in golang, a Component of the game interface
in Vue3/TypeScript, or a protocol representation of the game state and encoding
for the client's (GDL-formatted or binary) play/movement actions.

This README will be filled out as the compiler's functionality and flags expands.