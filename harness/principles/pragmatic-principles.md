# Pragmatic Principles

Four guidelines articulated by Andrew Hunt and David Thomas in *The Pragmatic Programmer* (1999; 20th Anniversary Edition, 2019). Each is a heuristic about *change*: code spends most of its life being modified, and these principles bias the work toward modifications that are cheap, local, and safe.

## DRY — Don't Repeat Yourself

> Every piece of knowledge must have a single, unambiguous, authoritative representation within a system.

DRY is about *knowledge*, not characters. Two functions that contain the same loop are not necessarily a DRY violation if the loops embody different rules that happen to look alike today. Two functions that encode the same business rule are a DRY violation even if they share no syntax. The test is: *if this rule changes, how many places must change in lockstep?* If the answer is more than one, the rule has more than one home.

Hunt and Thomas also coined a less-quoted but equally important warning: *imposed duplication* (forced by tooling), *inadvertent duplication* (one developer doesn't know about the other's code), *impatient duplication* (a shortcut), and *interdeveloper duplication* (teams don't talk). Each has a different remedy.

## Orthogonality

Two or more things are orthogonal if changes in one do not affect any of the others. An orthogonal system is one where a change in one place produces a change in exactly one place. This is the same force named in [[coupling-cohesion]] (loose coupling, high cohesion) and [[separation-of-concerns]] — *orthogonality* is the property the system has when those principles are honored.

The test: estimate the number of files affected by a typical change. The smaller the blast radius, the more orthogonal the system.

## ETC — Easier to Change

The meta-principle from the 20th Anniversary Edition: when choosing between two designs, prefer the one that leaves the code **easier to change**. Every other pragmatic principle reduces to a special case of ETC.

ETC is a *value*, not a rule. It does not prescribe; it provides the tiebreaker. When DRY and YAGNI seem to conflict, ETC settles it: which design will the next change be cheaper against?

## Broken Windows

> Don't live with broken windows.

A metaphor borrowed from criminology: a building with one broken window left unrepaired signals that no one cares, and the rest of the windows soon break too. In code: one tolerated smell — a TODO ignored, a test commented out, a known bug left in the tracker — lowers the standard for everything around it.

The remedy is not heroic cleanup; it is the *first* small fix. Board up the window. The signal that someone cares is more important than the size of the repair.

## What it asks of you

- When you write the same business rule in two places, ask which one is authoritative. If the answer is "both," DRY is being violated.
- When a change touches files in five directories, ask whether the system is orthogonal along the right axis. See [[separation-of-concerns]].
- When two designs are both correct, pick the one that the next change will be cheaper against. See [[simple-design]].
- When you notice a smell and step over it, you have just lowered the bar for everyone who follows. Fix it, file it, or remove the temptation.

## IRON LAW

**Knowledge has exactly one home.** If a business rule, a magic number, a configuration value, or an algorithm lives in two places, the system already lies — one of the two copies is wrong, you just don't know which yet.

## GOLDEN RULES

- **Aim for code where a typical change touches one place.** Measure your designs by blast radius, not by line count.
- **Aim for the *first* fix, not the *complete* fix.** A repaired window beats a planned renovation.
- **Aim for orthogonality between layers and between features.** A change to the payment provider should not touch the shopping cart UI.

## Anti-patterns

- A constant duplicated across config files because "it's just easier."
- A "shared" module that everything imports — see [[coupling-cohesion]].
- A change that requires editing the same regex in three places.
- A TODO that has survived two release cycles.
- A test directory containing `.skip` files left in for "now."

## References

- Hunt, A., & Thomas, D. (1999). *The Pragmatic Programmer: From Journeyman to Master*. Addison-Wesley.
- Hunt, A., & Thomas, D. (2019). *The Pragmatic Programmer: 20th Anniversary Edition*. Addison-Wesley. (Introduces ETC as the meta-principle.)
- Wilson, J. Q., & Kelling, G. L. (1982). "Broken Windows." *The Atlantic Monthly*. (Origin of the metaphor.)
