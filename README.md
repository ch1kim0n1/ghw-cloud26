# ghw-cloud26

## Overview
This repository contains the MVP documentation and implementation plan for a cloud-assisted contextual ad insertion system built during MLH Global Hack Week.

The product strategy is:
- Context-Aware Fused Ad Insertion (CAFAI)

The MVP goal is:
- analyze a provided H.264 MP4
- propose valid insertion slots automatically
- let the operator choose a slot and optionally edit the generated product line
- generate a short context-aware bridge clip
- stitch that clip into the source video with basic audio continuity
- export one downloadable preview MP4

## MVP Summary
The engineering-facing MVP contract is:
- supported source videos for the main MVP path are 10-20 minute H.264 MP4 files
- the system proposes the top 3 valid anchor-frame insertion slots when possible
- the operator can reject slots and request up to 2 re-picks
- CAFAI generation creates a 5-8 second bridge clip
- final preview download is served from local storage
- Azure Blob Storage is used only for temporary cloud artifacts during generation and rendering
- job states are coarse: `queued -> analyzing -> generating -> stitching -> completed|failed`

## Documentation
Core engineering docs live in [absolute-documents/01_Product_Design_Document.md](/C:/Users/Vladislav%20Kondratyev/Desktop/GitHub%20Repos/ghw-cloud26/absolute-documents/01_Product_Design_Document.md) through [absolute-documents/08_Task_Decomposition_Plan.md](/C:/Users/Vladislav%20Kondratyev/Desktop/GitHub%20Repos/ghw-cloud26/absolute-documents/08_Task_Decomposition_Plan.md).

Recommended reading order:
1. [01_Product_Design_Document.md](/C:/Users/Vladislav%20Kondratyev/Desktop/GitHub%20Repos/ghw-cloud26/absolute-documents/01_Product_Design_Document.md)
2. [02_System_Architecture_Document.md](/C:/Users/Vladislav%20Kondratyev/Desktop/GitHub%20Repos/ghw-cloud26/absolute-documents/02_System_Architecture_Document.md)
3. [03_Technical_Specifications.md](/C:/Users/Vladislav%20Kondratyev/Desktop/GitHub%20Repos/ghw-cloud26/absolute-documents/03_Technical_Specifications.md)
4. [06_API_Contracts.md](/C:/Users/Vladislav%20Kondratyev/Desktop/GitHub%20Repos/ghw-cloud26/absolute-documents/06_API_Contracts.md)
5. [07_Data_Schema_Definitions.md](/C:/Users/Vladislav%20Kondratyev/Desktop/GitHub%20Repos/ghw-cloud26/absolute-documents/07_Data_Schema_Definitions.md)
6. [08_Task_Decomposition_Plan.md](/C:/Users/Vladislav%20Kondratyev/Desktop/GitHub%20Repos/ghw-cloud26/absolute-documents/08_Task_Decomposition_Plan.md)

## Azure Service Choices
The current MVP stack assumes:
- analysis: Azure Video Indexer + Azure OpenAI
- CAFAI generation: Azure Machine Learning + Azure OpenAI
- audio generation and alignment: Azure AI Speech
- final render: Azure Container Apps running ffmpeg, with Azure Blob Storage for intermediate artifacts

## Artifact Flow
```text
generation output
      |
      v
Azure Blob Storage (temporary)
      |
      v
render worker pulls artifact
      |
      v
final preview written to Blob
      |
      v
preview copied back to local storage
      |
      v
download and debugging access
```
