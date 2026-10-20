# South Guangdong Cultural Tour Mini Program

This project is a WeChat Mini Program for the South Guangdong Cultural Tour AI System.

## Project Structure

- `miniprogram/`: The source code of the mini program.
  - `pages/`: Application pages.
    - `index/`: Home page.
    - `merchant/`: Merchant dashboard and tools.
    - `user/`: User-facing pages (Route, Map, Blog, Guide).
    - `admin/`: Admin pages (Knowledge Base).
  - `custom-tab-bar/`: Custom TabBar component using TDesign.
  - `app.json`: Global configuration.

## Setup Instructions

1. **Install Dependencies**:
   This project uses `tdesign-miniprogram`. You need to install it.
   ```bash
   npm install
   ```

2. **Build NPM**:
   Open WeChat DevTools.
   Go to **Tools** -> **Build npm**.
   This will create `miniprogram/miniprogram_npm/` directory containing TDesign components.

3. **Compile**:
   The project should compile successfully.

## Dependencies

- `tdesign-miniprogram`: UI Component Library.

## Notes

- **Mock Data**: The pages currently use static mock data.
- **Images**: Most images are using Unsplash or Wikimedia URLs. Ensure you have internet access in the simulator.
- **Map**: The map view uses a static image background. For real map functionality, replace the `<image>` with `<map>` component and configure map SDK.
- **TabBar**: A custom TabBar is implemented in `miniprogram/custom-tab-bar/`.
