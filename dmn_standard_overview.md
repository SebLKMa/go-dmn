# Decision Model and Notation (DMN) Standard Overview

The **Decision Model and Notation (DMN)** is a vendor-neutral, open standard managed by the [Object Management Group (OMG)](https://www.omg.org/spec/DMN/) for **modeling, describing, and executing operational business decisions and repeatable business rules**. 

DMN serves as a crucial bridge between business professionals and IT engineers. It provides a standardized visual language that domain experts can design and read, while remaining precise enough to be compiled directly into executable source code for decision automation engines.

---

## 🧱 The 3 Architectural Levels of DMN
The standard addresses decision-making by splitting the problem into three distinct visual and logical layers:

| Level | Component | Purpose |
| :--- | :--- | :--- |
| **L1** | **Business Process Model** | Usually built using [BPMN](https://www.bpmn.org/) (Business Process Model and Notation). It maps out the broader workflow and specifies the precise "Business Rule Tasks" where a decision is required. |
| **L2** | **Decision Requirements Diagram (DRD)** | The core layout of DMN. It visually maps out the decisions to be made, the data sources needed, and how they interrelate using standardized shapes. |
| **L3** | **Decision Logic** | The underlying mathematical or logical syntax (such as Decision Tables or Friendly Enough Expression Language—FEEL) that evaluates inputs to generate an output. |

---

## 🎨 Core Elements of a Decision Requirements Diagram (DRD)
A DRD visually structures complex business logic so anyone can trace how a decision is reached. It relies on four primary shapes:

* **Decision Nodes (Rectangles):** Represents the actual act of determining an outcome based on set rules.
* **Input Data Nodes (Ovals):** Information fed into the model from external systems, databases, or users.
* **Business Knowledge Models (Rectangles with clipped corners):** Reusable functions or sub-routines containing specific business rules (e.g., a tax rate calculator or a credit scoring formula).
* **Requirements (Arrows):** Connect the shapes to establish dependencies (e.g., specifying which Input Data flows into a specific Decision).

---

## ⚙️ How Decision Logic is Expressed (L3)
Behind every decision node lies the actual logic rule. DMN uses **Boxed Expressions** to display this:

* **Decision Tables:** The most widely used approach. A grid format where columns represent inputs/conditions and rows dictate the corresponding outputs in an intuitive "If-Then" structure.
* **FEEL (Friendly Enough Expression Language):** A specialized text-based language optimized for business professionals. It allows users to write complex logical, date, and math calculations without requiring full programming skills.
* **Contexts & Relations:** Used to define structural data maps, function definitions, lists, or custom evaluation tables.

---

## 🤝 The "Triple Crown" of Process Modeling
DMN is rarely used completely in isolation. It belongs to a trio of complementary OMG standards known as the **Triple Crown of process modeling**, designed to map any organizational operations:
1. **BPMN (Processes):** Maps sequential, prescriptive workflows (*"What do we do next?"*).
2. **CMMN (Cases):** Maps reactive, non-linear ad-hoc case management (*"How do we handle this unpredictable scenario?"*).
3. **DMN (Decisions):** Extracts the heavy logic rules from processes and maps out specific calculations (*"What is the final evaluation criteria?"*).

---

## 🚀 Common Tools & Automations
DMN files can be natively deployed on multiple enterprise automation engines such as [Camunda](https://camunda.com/), [Red Hat Decision Manager (Kogito/Drools)](https://drools.org/), and Trisotech.
