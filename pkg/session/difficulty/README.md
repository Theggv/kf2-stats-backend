# Difficulty Calculation

## Wave Difficulty

```math

\begin{aligned}
& score = zeds\_difficulty * wave\_size\_penalty * total\_players\_penalty * kiting\_penalty, \\

& zeds\_difficulty = \dfrac{\sum_{i=1}^{zed\_types} count_i * weight_i}{total\_zeds}, \\

\end{aligned}
```

TBA

## Session Difficulty

TBA

### Examples

| Settings | avg_zeds_diff | score | score^2 |
| -------- | ------------- | ----- | ------- |
| Hard - Long 		| 1.09 | 1.45  | 2.11
| Suicidal - Long 	| 1.41 | 1.47  | 2.17
| Hoe - Long  		| 1.86 | 1.93  | 3.73
| #6 (13% larges)  	| 2.20 | 2.18  | 4.76
| #8 (19% larges)  	| 2.55 | 2.35  | 5.53
| #8 (23% larges)  	| 2.65 | 2.59  | 6.74
| #8 (27% larges)  	| 2.85 | 2.77  | 7.72
| miasma ts_mig_v3 HZ  | 3.93 | 4.12  | 17.02
| biolab ts_mig_v3 NCZ | 4.25 | 4.86  | 23.70