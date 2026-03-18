# Copyright (c) 2024 Microsoft Corporation.
# Licensed under the MIT License

"""专门用于孤立实体关系提取的提示词模板"""

ISOLATED_ENTITY_RELATIONSHIP_EXTRACTION_PROMPT = """
-Goal-
Given an isolated entity and its associated text content, identify potential relationships between this entity and other entities mentioned in the text. Focus ONLY on creating new relationships, do NOT extract new entities.

-Important Rules-
1. ONLY create relationships for the given isolated entity - do not extract any new entities
2. Base all relationships strictly on the provided text content
3. The isolated entity must be either source_entity or target_entity in every relationship
4. If no valid relationships can be found, output an empty list followed by {completion_delimiter}
5. Focus on factual, objective relationships supported by the text

-Candidate Entity List-
You are provided with a curated list of existing entity titles extracted from the current knowledge graph. When generating relationships, you MUST choose the other endpoint entity strictly from this list. If no entity in the list matches based on the text, output nothing.

Candidate Entities (one per line):
{candidate_entities}

-Steps-
1. Analyze the provided text to identify other entities that the isolated entity might be related to
2. For each potential relationship, extract:
   - source_entity: name of the source entity (must be the isolated entity or another entity from the text)
   - target_entity: name of the target entity (must be the isolated entity or another entity from the text)
   - relationship_description: explanation of the relationship based on the text
   - relationship_strength: numeric score (1-10) indicating relationship strength

Relationship Strength Guidelines:
- Strong (7-10): Direct causal relationships, clear associations
- Medium (4-6): Statistical correlations, indirect effects
- Weak (1-3): Potential associations, co-occurrence patterns

Format each relationship as ("relationship"{tuple_delimiter}<source_entity>{tuple_delimiter}<target_entity>{tuple_delimiter}<relationship_description>{tuple_delimiter}<relationship_strength>)

3. Return output as a list of relationships. Use **{record_delimiter}** as the list delimiter.

4. When finished, output {completion_delimiter}

######################
-Examples-
######################
Example 1:
Isolated Entity: AUTISM SPECTRUM DISORDER
Entity Description: A neurodevelopmental condition characterized by social communication difficulties and restricted interests
Text: Autism spectrum disorder affects approximately 1 in 54 children. Early intervention programs like ABA therapy can significantly improve outcomes for children with autism. Research shows that genetic factors play a significant role in autism development.
######################
Output:
("relationship"{tuple_delimiter}AUTISM SPECTRUM DISORDER{tuple_delimiter}ABA THERAPY{tuple_delimiter}ABA therapy is an early intervention program that can significantly improve outcomes for children with autism{tuple_delimiter}8)
{record_delimiter}
("relationship"{tuple_delimiter}AUTISM SPECTRUM DISORDER{tuple_delimiter}GENETIC FACTORS{tuple_delimiter}Research shows that genetic factors play a significant role in autism development{tuple_delimiter}9)
{completion_delimiter}

Example 2:
Isolated Entity: SEROTONIN
Entity Description: A neurotransmitter involved in mood regulation and social behavior
Text: Serotonin levels are often abnormal in individuals with autism. Selective serotonin reuptake inhibitors (SSRIs) are sometimes prescribed to manage anxiety symptoms in autism. The serotonin transporter gene has been linked to autism susceptibility.
######################
Output:
("relationship"{tuple_delimiter}SEROTONIN{tuple_delimiter}AUTISM{tuple_delimiter}Serotonin levels are often abnormal in individuals with autism{tuple_delimiter}8)
{record_delimiter}
("relationship"{tuple_delimiter}SEROTONIN{tuple_delimiter}SSRIs{tuple_delimiter}Selective serotonin reuptake inhibitors are sometimes prescribed to manage anxiety symptoms in autism{tuple_delimiter}7)
{record_delimiter}
("relationship"{tuple_delimiter}SEROTONIN{tuple_delimiter}SEROTONIN TRANSPORTER GENE{tuple_delimiter}The serotonin transporter gene has been linked to autism susceptibility{tuple_delimiter}8)
{completion_delimiter}

######################
-Real Data-
######################
Isolated Entity: {isolated_entity}
Entity Description: {entity_description}
Text: {input_text}
######################
Output:"""

CONTINUE_ISOLATED_RELATIONSHIP_PROMPT = "Continue analyzing the text to find additional relationships for the isolated entity. Add any new relationships below using the same format:\n"
LOOP_ISOLATED_RELATIONSHIP_PROMPT = "Are there any more relationships that can be identified for this isolated entity? Answer Y if there are more relationships to add, or N if there are none. Please answer with a single letter Y or N.\n"
