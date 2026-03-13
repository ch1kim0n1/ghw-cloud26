# Azure Placeholder Notes

Phase 0 does not call Azure services. The backend loads optional Azure-related
environment variables now so later phases can replace the no-op clients without
changing the startup shape.

Expected service areas for later phases:
- Azure Video Indexer for video analysis
- Azure OpenAI for narrative and prompt assistance
- Azure Machine Learning for CAFAI clip generation
- Azure AI Speech for spoken line synthesis and alignment
- Azure Blob Storage for temporary artifacts
- Azure Container Apps for ffmpeg-based preview rendering

Expected environment variables:
- `AZURE_VIDEO_INDEXER_URL`
- `AZURE_OPENAI_URL`
- `AZURE_ML_URL`
- `AZURE_SPEECH_URL`
- `AZURE_BLOB_URL`
- `AZURE_RENDER_URL`
