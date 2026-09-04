# Vinyl Catalog API - Bruno Collection

Bruno collection (OpenCollection YAML `.yml` format, Bruno v3.1+) for testing the Vinyl Catalog API locally and in production.

## Installation

1. Download and install Bruno from https://www.usebruno.com/
2. Clone this repository
3. Open Bruno and import the collection from the `bruno-collection` folder

## Setup

### Environment Variables

The collection includes two environments:

**Local (for local development):**
- `base_url`: `http://localhost:8787`
- `api_version`: `v1`

**Production (for deployed API):**
- `base_url`: `https://vinyl-catalog-api.nataliagranato.xyz`
- `api_version`: `v1`

### Authentication

Most endpoints require JWT authentication. Follow these steps:

1. Run the **Login** request in the Authentication folder
2. Copy the `token` from the response
3. Set the `jwt_token` environment variable in Bruno:
   - Click on the environment selector (top right)
   - Select the environment (Local or Production)
   - Add a new variable: `jwt_token` with the token value
4. All authenticated requests will now use this token

## Collection Structure

### System (Public Endpoints)
- **Get API Root** - Returns the API documentation page
- **Health Check** - Verifies API, database, and storage connectivity
- **Get Metrics** - Returns database and storage statistics

### Authentication
- **Login** - Authenticate and get JWT token

### Vinyls (Requires Auth)
- **Get All Vinyls** - List all vinyl records
- **Create Vinyl** - Create a new vinyl record
- **Get Vinyl by ID** - Retrieve specific vinyl
- **Update Vinyl** - Update existing vinyl
- **Delete Vinyl** - Remove vinyl from catalog
- **Toggle Favorite** - Toggle favorite status

### Tracks (Requires Auth)
- **Get Tracks** - List tracks for a vinyl
- **Create Track** - Add new track to vinyl
- **Update Track** - Update existing track
- **Delete Track** - Remove track from vinyl

### Profile
- **Get Profile** - Get public profile (Public)
- **Update Profile** - Update user profile (Requires Auth)
- **Upload Profile Photo** - Upload profile photo (Requires Auth)

### Uploads (Requires Auth)
- **Upload Cover Image** - Upload vinyl cover image
- **Get Uploaded File** - Serve uploaded files (Public)

## Quick Start

1. **Import Collection**
   ```
   File → Import → Import Collection → Select bru-collection folder
   ```

2. **Set Environment**
   - Select "Local" for local testing
   - Select "Production" for deployed API testing

3. **Authenticate**
   - Run the **Login** request
   - Copy the token from response
   - Set `jwt_token` environment variable

4. **Test Endpoints**
   - Start with public endpoints (System, Get Profile)
   - Then test authenticated endpoints (Vinyls, Tracks, Profile, Uploads)

## Local Development

To test locally:

1. Start the local development server:
   ```bash
   cd /home/nataliagranato/vinyl-catalog
   npm run dev
   ```

2. Select "Local" environment in Bruno

3. Run the Health Check endpoint to verify connectivity

## Production Testing

To test the deployed API:

1. Select "Production" environment in Bruno

2. Run the Health Check endpoint to verify connectivity

3. Authenticate with Login endpoint

4. Test all endpoints

## Tips

- Use the collection runner to run multiple requests in sequence
- Set up request sequences for complex workflows (create vinyl → add tracks → upload cover)
- Use the environment variables to easily switch between local and production
- The collection includes test assertions to verify responses

## File Uploads

For file upload requests (cover images, profile photos):

1. Update the file path in the request body
2. Ensure the file exists locally
3. Run the request to upload

## Troubleshooting

**401 Unauthorized:**
- Check that `jwt_token` environment variable is set
- Verify the token is not expired (24h default)
- Run Login again to get a fresh token

**404 Not Found:**
- Verify the vinyl ID or track ID is correct
- Check that the resource exists in the database

**Connection Errors:**
- Ensure the API is running (local: `npm run dev`)
- Check the base_url is correct for the selected environment
- Verify network connectivity

## Credits

Built with [Bruno](https://www.usebruno.com/) - Fast and open-source API client.