**Note: This document is licensed under [Creative Commons Attribution Share Alike 4.0](http://creativecommons.org/licenses/by-sa/4.0/).**

Since its introduction in the 1970s, the CHIP-8 programming language has spawned a variety of dialects and descendant languages. This document aims to briefly describe the various programming languages related to CHIP-8. Entries are listed chronologically.

This document should not be considered comprehensive; certain CHIP-8 descendants may be missing. To suggest a revision, please [open an issue](https://github.com/mattmikolay/chip-8/issues).

## CHIP-8C (RCA, 1978)
Described as “the color-language addition to CHIP-8,” CHIP-8C promised control of three background colors and eight foreground colors in conjunction with RCA's VIP Color Board. It is not known whether this language was released publicly, or if it eventually grew into RCA's CHIP-8X. [\[9\]](#references)

## CHIP-8I (Rick Simpson, 1978)
A modification to the original CHIP-8 interpreter that provides three new instructions to support hardware I/O.
 [\[10\]](#references)

## CHIP-10 (Ben H. Hutchinson, Jr., 1979)
A modified version of CHIP-8 providing an expanded screen resolution of 128 x 64. [\[8\]](#references)

## CHIP-8 II (Tom Swan, 1979)
A modification to the original CHIP-8 interpreter, and an optional extension to Rick Simpson's CHIP-8I language, that adds another instruction to read from the VIP's input port. This change was made to allow for the creation of two-player games controlled by an ASCII keyboard. [\[12\]](#references)

## HI-RES CHIP-8 (Tom Swan, 1980)
A modified version of CHIP-8 providing an expanded screen resolution of 64 x 128 and increased speed. [\[11\]](#references)

## CHIP-8III (John Chmielewski, 1980)
A modified version of CHIP-8 aimed at providing the functionality of both Rick Simpson's CHIP-8I and Tom Swan's CHIP-8 II while maintaining compatibility with the original CHIP-8 language. [\[3\]](#references)

## CHIP-8E (Gilles Detillieux, 1980)
A rewrite of the original CHIP-8 interpreter that adds fourteen additional instructions and support for hardware I/O. [\[4\]](#references)

## CHIP-8X (RCA, 1980)
An “expanded version of the original CHIP-8 interpreter” meant for use with a series of expansion modules marketed by RCA for the COSMAC VIP system: the VP-590 Color board, the VP-595 Simple Sound board, and the VP-580 Expansion Keypad. The CHIP-8X language adds color graphics, extended sound capabilities, and support for a second keypad to the original CHIP-8 language. [\[6, 13\]](#references)

## CHIP-8Y (Bob Casey, 1981)
A modified version of CHIP-8 supporting hardware I/O while maintaining compatibility with the original CHIP-8 language. [\[2\]](#references)

## CHIP-8M (Steven V. Gunhouse, 1982 – 1983)
A modified version of CHIP-8 providing instructions to output International Morse Code tones. [\[7\]](#references)

## CHIP-8 AE (Mike E. Franklin, Tony Hill, Larry Owen, 1984)
A modified version of CHIP-8 for both the CDP1861 PIXIE chip and the ACE VDU display board. This variant provides lo-res, med-res, and hi-res graphic modes; 16 banks of 16 variables each; 16 timers; ASCII keyboard support; and ASCII character output. [\[16, 17\]](#references)

## SUPER-CHIP (Erik Bryntse, 1991)
An expansion of the CHIP-8 instruction set first introduced as a modification of Andreas Gustafsson's CHIP-8 emulator for the Hewlett-Packard HP 48 series of calculators. This CHIP-8 variant offers a higher screen resolution, larger sprites, larger fonts, screen scrolling, and more. It is also known as “S-CHIP”. [\[1\]](#references)

## MEGA-CHIP8 (Martijn Wenting, 2007)
A modern extension to the CHIP-8 language providing an expanded screen resolution, larger sprites, color graphics, and digital sound. [\[14, 15\]](#references)

## XO-CHIP (John Earnest, 2014)
A modern extension to the CHIP-8 language introduced alongside the [Octo](https://github.com/JohnEarnest/Octo) IDE. This variant supports additional save and load instructions, color graphics, programmable audio, and screen scrolling. [\[5\]](#references)

# References
[1]: Erik Bryntse. *SUPER-CHIP v1.1 (now with scrolling)*. May 28, 1991. URL: http://devernay.free.fr/hacks/chip8/schip.txt (visited on 12/21/2015).

[2]: Bob Casey. “CHIP-8 with I/O Modifications: CHIP-8Y”. In: *VIPER* 3.1 (Apr.–May 1981), p. 18.

[3]: John Chmielewski. “CHIP-8III”. In: *VIPER* 2.7 (Feb. 1980), pp. 6–7.

[4]: Gilles Detillieux. “CHIP-8E”. In: *VIPER* 2.8/9 (Mar.–Apr. 1980), pp. 15–17.

[5]: John Earnest. *Octo Extensions*. Sept. 19, 2015. URL: https://github.com/JohnEarnest/Octo/blob/gh-pages/docs/XO-ChipSpecification.md (visited on 12/21/2015).

[6]: *Game Manual II*. RCA Corporation, 1980.

[7]: Steven Vincent Gunhouse. “Special Routines for Morse Code from Standard CHIP-8”. In: *VIPER* 4.5 (Dec. 1982–Jan. 1983), pp. 2–4.

[8]: Ben H. Hutchinson. “CHIP-10 INTERPRETER FOR THE COSMAC VIP”. In: *VIPER* 1.7 (Feb. 1979), pp. 11–16.

[9]: “NEW FROM RCA”. In: *VIPER* 1.2 (Aug. 1978), p. 6.

[10]: Rick Simpson. “A Modification of CHIP-8 to Provide I/O Instructions”. In: *VIPER* 1.3 (Sept. 1978), p. 4.

[11]: Tom Swan. “HI-RES CHIP-8”. In: *VIPER* 2.6 (Jan. 1980), pp. 4–10.

[12]: Tom Swan. “KEYBOARD KONTROL”. In: *VIPER* 2.4 (Oct. 1979), pp. 5–9.

[13]: *VP580, VP585, VP590, VP595 Instruction Manual*. RCA Corporation.

[14]: Martijn Wenting. *MEGACHIP8 DEVKIT (CHIP-8/SUPERCHIP, MEGACHIP8, 2007)*. URL: http://www.revival-studios.com/other.php (visited on 12/21/2015).

[15]: Martijn Wenting. *MEGA-CHIP8 v1.0b*. URL: http://github.com/gcsmith/gchip/blob/master/docs/megachip10.txt (visited on 12/21/2015).

[16]: “CHIP-8 AE (ACE EXTENDED)”. In: *Ipso Facto* 40 (May 1984), pp. 12–29.

[17]: M.E. Franklin. “CHIP 8 AE Disassembler”. In: *Ipso Facto* 41 (July 1984), pp. 37–40.
