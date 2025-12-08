# Tree of Thoughts Example

## Overview

This example demonstrates the **Tree of Thoughts (ToT)** pattern, a search-based reasoning framework where problem-solving is modeled as a search through a tree structure. ToT extends linear chain-of-thought to multi-path exploratory search, generating multiple candidate "thoughts" at each step, evaluating their feasibility, and pruning unpromising branches while expanding the most promising ones.

## What is Tree of Thoughts?

Tree of Thoughts is an agentic architecture that implements systematic exploration of the problem space through a search tree. Unlike linear reasoning approaches, ToT explores multiple solution paths simultaneously and uses evaluation to guide the search toward the goal.

## The Five Key Phases

1. **Decomposition**: Break down the problem into a series of steps or thoughts
2. **Thought Generation**: Generate multiple potential next steps, creating branches in the search tree
3. **State Evaluation**: Evaluate each thought for:
   - Validity: Does this move follow the rules?
   - Progress: Are we getting closer to the solution?
   - Heuristics: How promising is this path?
4. **Pruning & Expansion**: Prune invalid or unpromising branches, continue from the most promising active branches
5. **Solution**: Continue until reaching a goal state; solution is the path from root to goal

## Architecture

```
                Problem Start State
                        │
        ┌───────────────┼���──────────────┐
        │               │               │
    Thought 1       Thought 2       Thought 3
        │               │               │
    ┌───┼───┐       ┌───┼──���┐       [Pruned]
    │       │       │       │
 State   State   State   State
  1.1     1.2     2.1     2.2
    │       X       │       X
    │   [Invalid]   │   [Low Score]
 Goal               │
[Found!]        Continue...
```

## Core Components

### 1. ThoughtState Interface

Represents a state in the search tree:

```go
type ThoughtState interface {
    IsValid() bool            // Check if state follows rules
    IsGoal() bool            // Check if this is the solution
    GetDescription() string  // Human-readable description
    Hash() string           // Unique identifier for cycle detection
}
```

### 2. ThoughtGenerator Interface

Generates possible next states:

```go
type ThoughtGenerator interface {
    Generate(ctx context.Context, current ThoughtState) ([]ThoughtState, error)
}
```

### 3. ThoughtEvaluator Interface

Evaluates state quality:

```go
type ThoughtEvaluator interface {
    Evaluate(ctx context.Context, state ThoughtState, pathLength int) (float64, error)
}
```

## Configuration

```go
type TreeOfThoughtsConfig struct {
    Generator    ThoughtGenerator  // Creates new states
    Evaluator    ThoughtEvaluator // Scores states
    MaxDepth     int              // Maximum search depth (default: 10)
    MaxPaths     int              // Max active paths to maintain (default: 5)
    Verbose      bool             // Enable detailed logging
    InitialState ThoughtState     // Starting state
}
```

## Example: River Crossing Puzzle

The included example solves the classic wolf-goat-cabbage river crossing puzzle:

**Problem**: A farmer needs to transport a wolf, a goat, and a cabbage across a river.

**Constraints**:
1. The boat can only carry the farmer and at most one other item
2. The wolf cannot be left alone with the goat (wolf eats goat)
3. The goat cannot be left alone with the cabbage (goat eats cabbage)

### Solution

The Tree of Thoughts agent finds the solution through systematic search:

```
Step 1: Farmer takes Goat across
Step 2: Farmer returns alone
Step 3: Farmer takes Wolf across
Step 4: Farmer returns with Goat
Step 5: Farmer takes Cabbage across
Step 6: Farmer returns alone
Step 7: Farmer takes Goat across
```

## Running the Example

```bash
cd examples/tree_of_thoughts
go run main.go
```

## How It Works

1. **Initialization**: Starts with all items on the left bank
2. **Expansion**: Generates all possible legal moves (farmer alone, farmer + wolf, farmer + goat, farmer + cabbage)
3. **Validation**: Checks each state for rule violations (e.g., wolf and goat alone together)
4. **Cycle Detection**: Uses state hashing to avoid revisiting the same state
5. **Evaluation**: Scores states based on progress (how many items on right bank)
6. **Pruning**: Removes invalid states and keeps only the top N most promising paths
7. **Goal Check**: Continues until finding a state with all items on the right bank

## When to Use Tree of Thoughts

**Best suited for**:
- Logic puzzles with clear rules and goal states
- Complex planning problems where action order and constraints are critical
- Problems requiring exploration of multiple strategies
- Constraint satisfaction problems

**Not ideal for**:
- Simple, straightforward tasks
- Problems without clear validation rules
- Tasks where evaluation is subjective or ambiguous

## Advantages

- **Robustness**: Systematic exploration dramatically reduces errors compared to single-pass approaches
- **Correctness**: Reliability comes from algorithmic soundness, not model memorization
- **Combinatorial Complexity**: Excels at problems with large possibility spaces
- **Verifiable**: Each step can be validated against rules

## Disadvantages

- **Computational Cost**: Requires many more LLM calls and state management operations
- **Speed**: Slower than linear approaches due to exploration
- **Evaluator Dependency**: Search effectiveness heavily depends on evaluation function quality

## Key Insight

Unlike chain-of-thought prompting which relies on the LLM's memorized knowledge, Tree of Thoughts **discovers** solutions through verifiable search. Even if the LLM makes a suboptimal choice at one branch, the agent continues exploring other branches, guaranteeing eventual correctness for problems with well-defined rules.

## Implementation Notes

The example demonstrates:
- Custom state representation (`RiverState`)
- Rule-based validation (checking for forbidden combinations)
- State generation (all possible boat crossings)
- Heuristic evaluation (progress toward goal)
- Cycle detection (avoiding revisited states)

## References

- [All Agentic Architectures](https://github.com/FareedKhan-dev/all-agentic-architectures) - Original Tree of Thoughts implementation
- Pattern #9: Tree of Thoughts (ToT)
- [Tree of Thoughts Paper](https://arxiv.org/abs/2305.10601) - "Tree of Thoughts: Deliberate Problem Solving with Large Language Models"

## License

This implementation is part of the langgraphgo project.
